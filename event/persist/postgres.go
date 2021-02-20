package persist

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres" //importing postgres driver
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// NewPostgresRepository Creates a new content/listing repository
// backed by PostGreSQL
func NewPostgresRepository(db *sql.DB) PostgresRepository {
	logger := logging.GetLogger()
	return PostgresRepository{sqlEngine: sqlx.NewDb(db, "postgres"), db: db, Logger: logger.WithField("module", "persist.PostgresRepository")}
}

// PostgresRepository Repository backs by PostGreSQL
type PostgresRepository struct {
	sqlEngine *sqlx.DB
	db        *sql.DB
	Logger    *logrus.Entry
	mu        sync.Mutex
	lLock     sync.Mutex
}

func (pgr *PostgresRepository) GetListingsDetails(ids []string) []*models.SimpleListing {
	listings := []*models.SimpleListing{}
	err := pgr.sqlEngine.Select(&listings, `SELECT 
	listing.id as listing_id ,
	(value ->> 'id')::text as id ,
		value ->> 'site' AS site ,
		COALESCE(value ->> 'address_text', '') AS address ,
		COALESCE((value ->> 'floors')::integer,-1)  AS floors,
		COALESCE((value ->> 'beds')::integer,-1) AS beds ,
		COALESCE((value -> 'price'->> 'amount_week')::integer,-1) AS ppw,
		COALESCE((value -> 'price'->> 'currency'),'GBP') AS currency,
		COALESCE((value ->> 'baths')::integer,-1) AS baths,
		COALESCE(( value -> 'floorArea' ->> 'value')::integer,-1) AS floor_area,
		COALESCE(( value -> 'location' ->> 'lon')::float,0) as longitude,
		COALESCE(( value -> 'location' ->> 'lat')::float,0) as latitude,
		COALESCE((value ->> 'underOffer' )::bool, false) AS is_under_offer,
		COALESCE((value ->> 'let' )::bool, false)  AS is_let,
		COALESCE((value ->> 'delisted' )::bool, false)  AS is_delisted,
		COALESCE((value ->> 'images')::json ->> 0, '') as images,
		(COALESCE(length(((value ->> 'floorPlans')::json  ->> 0)),0) > 0) as floor_plans,
		COALESCE((value -> 'price'->> 'amount')::integer,-1) AS ppm,
		COALESCE(( value -> 'agent' ),'{}')::json AS agent,
		COALESCE((value ->> 'type'),'') as property_type,
		COALESCE((value ->> 'url'),'') as url,
		COALESCE(value ->> 'features', '[]')::json as features
	FROM listing 
	WHERE  id = ANY($1) ORDER BY updated_date DESC`, pq.Array(ids))

	if err != nil {
		pgr.Logger.WithError(err).Errorf("Can't read listings for IDs %v  ", ids)
		return []*models.SimpleListing{}
	}

	return listings
}

func (pgr *PostgresRepository) GetLatestContentIdsBySite(site string, start time.Time, end time.Time) []string {
	var items []string
	var logger = pgr.Logger.WithField("site", site)
	rows, err := pgr.db.Query(`SELECT id FROM listing_content
	INNER JOIN ( SELECT max(date) AS maxDate,value ->> 'listingId' AS listingId FROM listing_content 
	WHERE 
		value  -> 'session' ->> 'site'  = $1 AND  date BETWEEN $2 AND $3
	GROUP BY value ->> 'listingId') myquery 
	ON 
		myquery.maxDate = listing_content.date 
	AND 
		listing_content.value ->> 'listingId' =  myquery.listingId
		
		`, site, start, end)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			if err != nil {
				break
			}
			items = append(items, id)
		}
	}

	if err != nil {
		logger.WithError(err).Errorf("Can't read content")
	}
	return items
}

func (pgr *PostgresRepository) GetContentByIds(ids []string) []models.ParsedItem {
	var items []models.ParsedItem
	var logger = pgr.Logger.WithField("ids", ids)
	rows, err := pgr.db.Query(`SELECT id,type,content,value FROM listing_content  WHERE id = ANY($1)`, pq.Array(ids))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, eventType string
			var rawContent, value []byte
			err = rows.Scan(&id, &eventType, &rawContent, &value)
			if err != nil {
				break
			}

			var item models.ParsedItem
			err = json.Unmarshal(value, &item)
			if err != nil {
				break
			}

			content, err := base64.StdEncoding.DecodeString(string(rawContent))
			if err != nil {
				break
			}

			item.Content = content
			items = append(items, item)
		}
	}

	if err != nil {
		logger.WithError(err).Errorf("Can't read content")
	}
	return items
}

func (pgr *PostgresRepository) GetLastContentsByArea(areaCode string, site string) []models.PortalFetch {
	var items []models.PortalFetch
	var logger = pgr.Logger.WithField("site", site).WithField("area", areaCode)
	rows, err := pgr.db.Query(`SELECT
    value -> 'session' ->> 'site' as site,
    value -> 'session' ->> 'area' as area,
    value -> 'source' -> 'origin' ->> 'topic' as topic,
    (value -> 'source' ->> 'compressed') as compressed,
		content.content
FROM listing_content as content
INNER JOIN
(SELECT
	value -> 'session' ->> 'sessionId'  as sessionId
        FROM
            listing_content
    WHERE value -> 'session' ->> 'area' = $1 AND  value -> 'session' ->> 'site' = $2
ORDER BY date desc
LIMIT 1) as query ON query.sessionId = content.value -> 'session' ->> 'sessionId'
WHERE value -> 'source' ->> 'compressed' is not null and value -> 'source' -> 'origin' ->> 'topic' is not null`, areaCode, site)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var site, area, topic, rawCompressed sql.NullString
			var rawContent sql.NullString

			err = rows.Scan(&site, &area, &topic, &rawCompressed, &rawContent)
			if err == nil {
				now := time.Now()
				var sessionID = utils.NewUUID()
				var compressed bool
				var content []byte

				compressed, err = strconv.ParseBool(rawCompressed.String)
				if err == nil {
					content, err = base64.StdEncoding.DecodeString(rawContent.String)
					if err == nil {
						fetch := models.PortalFetch{
							Session:    models.Session{SessionID: sessionID, Site: site.String, Area: area.String},
							Origin:     models.Origin{Topic: topic.String, CreatedDate: &now},
							Compressed: compressed,
							Content:    content,
							Type:       models.SearchResultItemParse,
						}
						items = append(items, fetch)
					}
				}
			}
		}
	}
	if err != nil {
		logger.WithError(err).Errorf("Can't read content")
	}
	return items
}

func (pgr *PostgresRepository) GetListingWithImages(site string) map[string]string {
	items := make(map[string]string)
	var logger = pgr.Logger.WithField("site", site)
	rows, err := pgr.db.Query(`select value ->> 'id' from listing where value ->> 'site' = $1  AND value ->> 'images' is not null`, site)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id sql.NullString
			err = rows.Scan(&id)
			if err == nil {
				items[id.String] = id.String
			}
		}
	}
	if err != nil {
		logger.WithError(err).Errorf("Can't read content")
	}
	return items
}

func (pgr *PostgresRepository) SaveSource(content []*models.ParsedItem) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		var stmt *sql.Stmt
		stmt, err = tr.Prepare("INSERT INTO listing_content (id,date,value,type,content) VALUES ($1,$2,$3,$4,$5)")
		if err == nil {
			defer stmt.Close()
			for i := 0; i < len(content); i++ {
				parsedItem := content[i]

				if len(parsedItem.Source.Content) == 0 {
					return fmt.Errorf("No source for Parsed Item Site %s ListingID %s", parsedItem.Session.Site, parsedItem.ListingID)
				}

				var logger = pgr.Logger.WithField("site", parsedItem.Session.Site).WithField("listingId", parsedItem.ListingID)
				encodedContent := base64.StdEncoding.EncodeToString(parsedItem.Source.Content)

				// clear content
				parsedItem.Content = []byte{}        // not needed saved into listing event
				parsedItem.Source.Content = []byte{} // stored as a field to only keep a light jsonb

				var bytes []byte
				bytes, err = json.Marshal(&parsedItem)
				if err != nil {
					logger.Errorln("can't encode listing")
				} else {
					_, err = stmt.Exec(utils.NewUUID(), time.Now(), string(bytes), parsedItem.Type, encodedContent)
					if err != nil {
						tr.Rollback()
						logger.Errorln(err)
						return
					}
				}
			}
			tr.Commit()
		} else {
			pgr.Logger.Errorln("Can't create prepared statement", err)
			tr.Rollback()
		}
	} else {
		pgr.Logger.WithError(err).Errorln("can't save")
	}
	return
}

func (pgr *PostgresRepository) UpdateConfig(code string, content []byte) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		_, err = tr.Exec("UPDATE site SET site_config_draft = $1 WHERE code = $2", content, code)
		if err == nil {
			tr.Commit()
			return
		}
	}

	if tr != nil {
		tr.Rollback()
	}

	pgr.Logger.WithError(err).Errorf("can't save config for %s", code)
	return
}

func (pgr *PostgresRepository) PublishConfig(code string) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		ts := time.Now()
		_, err = tr.Exec("UPDATE site SET site_config  = site_config_draft, updateddate = $1 WHERE code = $2", ts, code)
		if err == nil {
			_, err = tr.Exec("INSERT INTO site_config (id,site_code,config,created_date) SELECT $1,code,site_config_draft,updateddate FROM site WHERE code = $2 ", utils.NewUUID(), code)
		}

		if err == nil {
			tr.Commit()
			return
		}

	}

	if tr != nil {
		tr.Rollback()
	}

	pgr.Logger.WithError(err).Errorf("can't publish code for %s", code)
	return
}

// SaveContents Implementing ContentRepository SaveContents method
func (pgr *PostgresRepository) SaveContents(data []*models.PortalFetch) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		var stmt *sql.Stmt
		stmt, err = tr.Prepare("INSERT INTO listing_content (id,date,value,type,content) VALUES ($1,$2,$3,$4,$5)")
		if err == nil {
			defer stmt.Close()
			for i := 0; i < len(data); i++ {

				fetchedContent := data[i]

				encodedContent := ""

				var b bytes.Buffer
				gz, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
				_, errCompress := gz.Write(fetchedContent.Content)
				if errCompress == nil {
					gz.Flush()
					gz.Close()
					encodedContent = base64.StdEncoding.EncodeToString(b.Bytes())
				} else {
					pgr.Logger.WithError(errCompress).Error("Can't compress data")
					encodedContent = base64.StdEncoding.EncodeToString(fetchedContent.Content)
				}

				fetchedContent.Content = []byte{}

				var bytes []byte
				bytes, err = json.Marshal(&fetchedContent)
				if err != nil {
					pgr.Logger.Errorln("can't encode listing")
				} else {
					_, err = stmt.Exec(utils.NewUUID(), time.Now(), string(bytes), fetchedContent.Type, encodedContent)
					if err != nil {
						tr.Rollback()
						pgr.Logger.Errorln(err)
						return
					}
				}

			}
			tr.Commit()
		} else {
			pgr.Logger.WithError(err).Errorln("Can't create prepared statement to save content", err)
			tr.Rollback()
		}
	} else {
		pgr.Logger.WithError(err).Errorln("can't open a transaction", err)
	}
	return
}

//Unlock unlock scraping task
func (pgr *PostgresRepository) Unlock(sessionID string, site string) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		_, err = tr.Exec("UPDATE scrape_area SET locked = false  , lockeddate = null WHERE session_id = $1 and site = $2", sessionID, site)
		if err == nil {
			tr.Commit()
			return
		}
	}

	if tr != nil {
		tr.Rollback()
	}

	pgr.Logger.WithError(err).Errorln("can't save")
	return
}

//CleanEvents Remove event without any update on listing details
func (pgr *PostgresRepository) CleanEvents(sessionID string, site string) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		now := time.Now()
		_, err = tr.Exec("UPDATE scrape_area SET locked = false  , lockeddate = null, last_success_date = $1 WHERE session_id = $2 and site = $3", now, sessionID, site)
		if err == nil {
			tr.Commit()
			return
		}
	}
	if tr != nil {
		tr.Rollback()
	}
	pgr.Logger.WithError(err).
		WithField("site", site).
		WithField("sessionId", sessionID).Errorf("can't clean session %s", sessionID)
	return
}

//GetListingIds load all listing checksums
func (pgr *PostgresRepository) GetListingIds(handler StringHandler) {
	rows, err := pgr.db.Query("SELECT id, value ->> 'id' as listingId,value ->> 'site' as site FROM listing")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, site, listingID sql.NullString
			err = rows.Scan(&id, &listingID, &site)
			if err == nil {
				if !listingID.Valid || !site.Valid {
					pgr.Logger.Errorf("listing %s has invalid details", id.String)
				} else if handler(utils.GetMD5Hash(fmt.Sprintf("%s_%s", site.String, listingID.String))) {
					break
				}
			} else {
				break
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Errorf("can't load listings ids")
	}
}

func (pgr *PostgresRepository) LoadChecksums() []*EventChecksums {
	var allChecksums []*EventChecksums
	rows, err := pgr.db.Query("SELECT id , value ->> 'site' as site, value ->> 'id' as listingId,value ->> 'checksums' as checksums FROM listing")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, site, listingId sql.NullString
			var rawChecksums []byte
			err = rows.Scan(&id, &site, &listingId, &rawChecksums)
			if err == nil && site.Valid && listingId.Valid {
				var listingChecksums map[int]string
				if len(rawChecksums) > 0 {
					err = json.Unmarshal(rawChecksums, &listingChecksums)
					if err != nil {
						break
					}
				} else {
					listingChecksums = make(map[int]string)
				}
				key := models.ListingKey{
					ID:   listingId.String,
					Site: site.String,
				}
				allChecksums = append(allChecksums, NewEventChecksums(key.GetHash(), listingChecksums))
			} else {
				pgr.Logger.Warnf(" listing id %s for site %s is invalid", id.String, site.String)
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Errorf("can't load check sums")
	}

	return allChecksums
}

func (pgr *PostgresRepository) GetAllListings(fromDate time.Time, handler func(area *models.RawListing)) {
	rows, err := pgr.db.Query(sqlQueryAllListing, fromDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ID sql.NullString
			var floors, beds, ppw, baths, floorArea sql.NullInt64
			var longitude, latitude sql.NullFloat64
			var underOffer, let, delisted sql.NullBool
			var createdDate time.Time
			err = rows.Scan(&ID, &floors, &beds, &ppw, &baths, &floorArea, &longitude, &latitude, &underOffer, &let, &delisted, &createdDate)
			if err == nil {
				location := models.NewLocation(longitude.Float64, latitude.Float64)
				handler(&models.RawListing{
					Baths:       int32(baths.Int64),
					Beds:        int32(beds.Int64),
					ID:          ID.String,
					FloorArea:   int32(floorArea.Int64),
					Floors:      int32(floors.Int64),
					Ppw:         int32(ppw.Int64),
					Location:    &location,
					UnderOffer:  underOffer.Bool,
					Let:         let.Bool,
					Delisted:    delisted.Bool,
					CreatedDate: &createdDate,
				})
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Error("Can't read listings")
	}
}

func (pgr *PostgresRepository) GetListingPlaces(id string) []*models.Place {
	places := []*models.Place{}
	rows, err := pgr.db.Query(`SELECT fsar.id,fsar.name,fsar.type,lon, lat,rating,  ST_Distance(fsar.geom::geography, listing.geom::geography)  FROM listing 
	INNER JOIN fsar ON ST_DWithin(fsar.geom, listing.geom, 0.0055)
	WHERE listing.geom is not null  AND listing.id = $1 and rating >= 3 AND type in (7843,7840,1, 7838,4613,7844) 
	order by ST_Distance(fsar.geom::geography, listing.geom::geography)`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ID, name sql.NullString
			var longitude, latitude, distance sql.NullFloat64
			var placeType, rating sql.NullInt64
			err = rows.Scan(&ID, &name, &placeType, &longitude, &latitude, &rating, &distance)
			if err == nil {
				places = append(places, &models.Place{
					ID:       ID.String,
					Name:     name.String,
					Lon:      longitude.Float64,
					Lat:      latitude.Float64,
					Rating:   rating.Int64,
					Type:     placeType.Int64,
					Distance: distance.Float64,
				})
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Errorf("can't load check sums")
	}

	return places
}

func (pgr *PostgresRepository) GetPlaces(handler func(place *models.Place)) {
	rows, err := pgr.db.Query("select id, name, type, lon, lat, rating from fsar")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ID, name sql.NullString
			var longitude, latitude sql.NullFloat64
			var placeType, rating sql.NullInt64
			err = rows.Scan(&ID, &name, &placeType, &longitude, &latitude, &rating)
			if err == nil {
				handler(&models.Place{
					ID:     ID.String,
					Name:   name.String,
					Lon:    longitude.Float64,
					Lat:    latitude.Float64,
					Rating: rating.Int64,
					Type:   placeType.Int64,
				})
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Error("Can't read places")
	}
}

func (pgr *PostgresRepository) GetAreaListings(areaID int, handler func(area *models.RawListing)) {
	rows, err := pgr.db.Query(sqlQueryAreaListing, areaID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ID, site, address, image sql.NullString
			var floors, beds, ppw, baths, floorArea, floorPlan sql.NullInt64
			var longitude, latitude sql.NullFloat64
			var underOffer, let, delisted sql.NullBool
			err = rows.Scan(&ID, &site, &address, &floors,
				&beds, &ppw, &baths, &floorArea,
				&longitude, &latitude, &underOffer, &let, &delisted, &image, &floorPlan)
			if err == nil {
				location := models.NewLocation(longitude.Float64, latitude.Float64)
				handler(&models.RawListing{
					Address:    address.String,
					Baths:      int32(baths.Int64),
					Beds:       int32(beds.Int64),
					ID:         ID.String,
					FloorArea:  int32(floorArea.Int64),
					Floors:     int32(floors.Int64),
					Ppw:        int32(ppw.Int64),
					Location:   &location,
					Image:      image.String,
					FloorPlan:  floorPlan.Int64 > 0,
					UnderOffer: underOffer.Bool,
					Let:        let.Bool,
					Delisted:   delisted.Bool,
				})
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Errorf("Can't read listings for areaID %d", areaID)
	}
}

func (pgr *PostgresRepository) GetListingEvents(id models.ListingKey) []*models.ListingEvent {
	var items []*models.ListingEvent
	var logger = pgr.Logger.WithField("listingId", id.ID).WithField("site", id.Site)
	rows, err := pgr.db.Query(`select
			date,
			value ->> 'eventType' AS eventType ,
			value ->> 'site' AS site ,
			value ->> 'address_text' AS address ,
			value ->> 'floors'  AS floors,
			value ->> 'beds' AS beds ,
			value -> 'price'->> 'amount_week' AS ppw,
			value -> 'price'->> 'amount' AS ppm,
			value ->> 'baths' AS baths,
			value ->> 'reception' AS reception,
			value -> 'floorArea' ->> 'value' AS floorArea from listing_event where value ->> 'site' = $2 and  value ->> 'id' = $1 order by date asc`, id.ID, id.Site)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var createdDate pq.NullTime
			var eventType, site, address sql.NullString
			var floors, beds, ppw, ppm, baths, reception, floorArea int

			err = rows.Scan(&createdDate, &eventType, &site, &address,
				&floors, &beds, &ppw, &ppm, &baths, &reception, &floorArea)
			if err == nil {
				listingEvent := models.ListingEvent{
					CreatedDate: createdDate.Time,
					EventType:   eventType.String, Site: site.String, AddressText: address.String,
					Floors: floors, Beds: beds, PricePerWeek: ppw, PricePerMonth: ppm,
					Baths: baths, Reception: reception, FloorArea: floorArea,
				}
				items = append(items, &listingEvent)
			}
		}
	}
	if err != nil {
		logger.WithError(err).Errorf("Can't read listing  %s for site %s  ", id.ID, id.Site)
	}
	return items
}

func (pgr *PostgresRepository) UpdateLastViewed(ids []models.ListingKey) (err error) {
	pgr.mu.Lock()
	defer pgr.mu.Unlock()
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		var stmt *sql.Stmt
		stmt, err = tr.Prepare("UPDATE listing_views set lastviewed_date =  $1 WHERE id = $2")
		if err == nil {
			defer stmt.Close()
			for i := 0; i < len(ids); i++ {
				now := time.Now()
				l := ids[i]
				_, err = stmt.Exec(now, l.GetHash())
				if err != nil {
					tr.Rollback()
					pgr.Logger.WithError(err).Errorln("can't save listing viewlast")
					return
				}
			}
			tr.Commit()
		} else {
			pgr.Logger.WithError(err).Errorln("Can't create prepared statement")
			tr.Rollback()
		}
	} else {
		tr.Rollback()
		pgr.Logger.WithError(err).Errorln("can't save")
	}
	return
}

func (pgr *PostgresRepository) GetLastSeenOldListings(dayOld int, handler ListingIDHandler) {
	rows, err := pgr.db.Query(`
	SELECT listing.value ->> 'site' as site,listing.value ->> 'id' as id,listing.value ->> 'url' as url from listing_views 
	LEFT JOIN listing ON listing.id = listing_views.id
	WHERE 
		DATE_PART('day', current_date - COALESCE(listing_views.lastviewed_date,listing.created_date))  BETWEEN $1 AND $2
		AND listing.delisted = false
	order by listing_views.lastviewed_date desc
	`, dayOld, dayOld+2)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var site, listingID, URL sql.NullString
			err = rows.Scan(&site, &listingID, &URL)
			if err == nil {
				if !listingID.Valid || !site.Valid {
					pgr.Logger.Errorf("listing %s/%s has invalid details", site.String, listingID.String)
				} else if handler(models.ListingKey{
					Site: site.String,
					ID:   listingID.String,
					URL:  URL.String,
				}) {
					break
				}
			} else {
				break
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Errorf("can't load listings ids")
	}
	return
}

// SaveEvents Implementing ListingRepository SaveUpdates method
func (pgr *PostgresRepository) SaveEvents(data []*models.Listing) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		var stmt *sql.Stmt
		stmt, err = tr.Prepare("INSERT INTO listing_event  (id,date,value,geom) VALUES ($1,$2,$3,ST_GeomFromText($4,4326))")
		if err == nil {
			defer stmt.Close()
			for i := 0; i < len(data); i++ {
				var bytes []byte
				l := data[i]
				bytes, err = json.Marshal(&l)
				if err != nil {
					pgr.Logger.WithField("listingId", l.ID).WithField("site", l.Site).
						WithField("sessionId", l.SessionID).Errorln("can't encode listing")
				} else {
					_, err = stmt.Exec(utils.NewUUID(), time.Now(), string(bytes), fmt.Sprintf("POINT(%f %f)", l.Geo.Lon, l.Geo.Lat))
					if err != nil {
						tr.Rollback()
						pgr.Logger.Errorln(err)
						return
					}
				}

			}
			tr.Commit()
		} else {
			pgr.Logger.Errorln("Can't create prepared statement", err)
			tr.Rollback()
		}
	} else {
		pgr.Logger.Errorln("can't save", err)
	}
	return
}

func (pgr *PostgresRepository) CreateListingViews(listings []*models.Listing) (int, error) {

	con, err := pgr.db.Conn(context.Background())
	if err != nil {
		return -1, err
	}

	defer con.Close()
	now := time.Now()

	for i, l := range listings {
		logger := pgr.Logger.WithField("listingId", l.GetHash()).WithFields(logrus.Fields{"site": l.Site, "sessionId": l.SessionID})
		_, err := con.ExecContext(context.Background(), `INSERT INTO listing_views (id,visited_date ,lastviewed_date,site) VALUES ($1,$2 ,$3, $4 )`, l.GetHash(), now, now, l.Site)
		if err != nil {
			logger.WithError(err).Error("can't persist listing")
			return i - 1, err
		}
	}

	return len(listings), nil
}

func (pgr *PostgresRepository) CreateListing(listings []*models.Listing) (int, error) {

	values := sqlValues{}

	for _, l := range listings {

		logger := pgr.Logger.WithFields(logrus.Fields{
			"listingId": l.ID,
			"site":      l.Site,
			"sessionId": l.SessionID,
		})

		l.AddressText = models.ReNewLine.ReplaceAllString(l.AddressText, " ")
		bytes, err := json.Marshal(&l)
		if err != nil {
			logger.WithError(err).Error("can't encode listing")
			values.Args = append(values.Args, []interface{}{})
			values.Error = err
			break
		}

		now := time.Now()
		internalID := l.GetHash()

		values.Args = append(values.Args, []interface{}{internalID, string(bytes), fmt.Sprintf("POINT(%f %f)", l.Geo.Lon, l.Geo.Lat), now, now})

		l.InternalID = internalID
	}

	if len(values.Args) == 0 {
		return -1, fmt.Errorf("no arguments specified")
	}

	con, err := pgr.db.Conn(context.Background())
	if err != nil {
		return -1, err
	}

	defer con.Close()
	for i, arguments := range values.Args {

		if len(arguments) == 0 {
			return i, values.Error
		}

		_, err := con.ExecContext(context.Background(), `INSERT INTO listing (id,value,geom,created_date,updated_date) VALUES ($1,$2,ST_GeomFromText($3,4326),$4 ,$5)`, arguments...)
		if err != nil {
			pgr.Logger.WithError(err).Error("can't persist listing")
			return i - 1, err
		}

	}

	return len(listings), nil
}

func (pgr *PostgresRepository) GetListings(ids []string) []*models.Listing {
	var items []*models.Listing
	for i := 0; i < len(ids); i++ {
		var id = ids[i]

		var logger = pgr.Logger.WithField("listingId", id)

		rows, err := pgr.db.Query(`SELECT created_date,updated_date,visited_date,lastviewed_date,delisted_date,value
			 	FROM listing
				 LEFT JOIN listing_views ON listing_views.id = listing.id
				WHERE listing.id = $1`, id)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var data []byte

				var createdDate pq.NullTime
				var updatedDate pq.NullTime
				var visitedDate pq.NullTime
				var lastViewedDate pq.NullTime
				var lastDelistedDate pq.NullTime

				err = rows.Scan(&createdDate, &updatedDate, &visitedDate, &lastViewedDate, &lastDelistedDate, &data)
				if err == nil {
					listing := models.NewListing()
					err = json.Unmarshal(data, &listing)
					if err == nil {
						if createdDate.Valid {
							listing.CreatedDate = createdDate.Time
						}

						if updatedDate.Valid {
							listing.UpdatedDate = updatedDate.Time
						}

						if visitedDate.Valid {
							listing.VisitedDate = visitedDate.Time
						}

						if lastViewedDate.Valid {
							listing.LastViewedDate = lastViewedDate.Time
						}

						if lastDelistedDate.Valid {
							listing.LastDelistedDate = lastDelistedDate.Time
						}

						listing.InternalID = listing.GetHash()

						items = append(items, &listing)
					}
				} else {
					logger.WithError(err).Errorf("Can't read listing  %s ", id)
					return items
				}
			}
		}

		if err != nil {
			logger.WithError(err).Errorf("can't load listing %v", ids)
		}
	}
	return items
}

func (pgr *PostgresRepository) UpdateListings(data []*models.Listing) (err error) {
	pgr.lLock.Lock()
	defer pgr.lLock.Unlock()
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		var stmt *sql.Stmt
		stmt, err = tr.Prepare(`UPDATE listing SET value = $1, updated_date = $2,
				 geom = ST_GeomFromText($3,4326) , delisted = $4 WHERE (value ->> 'id') = $5
					AND  (value ->> 'site') = $6 `)

		if err == nil {
			for i := 0; i < len(data); i++ {
				var bytes []byte
				l := data[i]
				l.AddressText = models.ReNewLine.ReplaceAllString(l.AddressText, " ")
				logger := pgr.Logger.WithField("listingId", l.ID).WithField("site", l.Site).
					WithField("sessionId", l.SessionID)

				bytes, err = json.Marshal(&l)
				if err != nil {
					logger.WithError(err).Errorln("can't encode listing")
				} else {
					now := time.Now()
					_, err = stmt.Exec(string(bytes), now, fmt.Sprintf("POINT(%f %f)", l.Geo.Lon, l.Geo.Lat), l.Delisted, l.ID, l.Site)
					if err != nil {
						tr.Rollback()
						logger.WithError(err).Errorln(err)
						return
					}
				}
			}
		} else {
			pgr.Logger.WithError(err).Errorln("can't create a statement", err)
		}
		tr.Commit()
	} else {
		pgr.Logger.WithError(err).Errorln("can't create listings", err)
	}
	return
}

func (pgr *PostgresRepository) GetSitePortalConfigVersions(code string) []models.PortalConfig {
	return []models.PortalConfig{}
}

func (pgr *PostgresRepository) GetSites() []models.Site {
	logger := pgr.Logger
	items := []models.Site{}
	rows, err := pgr.db.Query("SELECT code,url,name,config as site_config,group_name,createdDate,updatedDate,enabled FROM site")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var data []byte
			var site models.Site
			err = rows.Scan(&site.Code, &site.URL, &site.Name, &data, &site.GroupName, &site.CreatedDate, &site.UpdatedDate, &site.Enabled)
			if err == nil {
				items = append(items, site)
			} else {
				logger.WithError(err).Errorln("Can't read sites ")
				return items
			}
		}
	} else {
		pgr.Logger.WithError(err).Errorln("can't GetSites", err)
	}
	return items
}

func (pgr *PostgresRepository) AddSite(siteDetails models.Site) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		ts := time.Now()
		_, err = tr.Exec(`INSERT INTO site (date, code, url, name, createddate, updateddate, group_name, enabled) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			ts, siteDetails.Code, siteDetails.URL, siteDetails.Name, ts, ts, siteDetails.GroupName, siteDetails.Enabled)

		if err == nil {
			tr.Commit()
			return
		}

	}

	if tr != nil {
		tr.Rollback()
	}

	pgr.Logger.WithError(err).Errorf("can't add site with code %s", siteDetails.Code)
	return
}

func (pgr *PostgresRepository) SearchSite(term string) []models.Site {
	logger := pgr.Logger
	items := []models.Site{}
	err := pgr.sqlEngine.Select(&items, `SELECT 
			code,url,name,enabled  FROM
			site WHERE lower(name) like $1 LIMIT 10`, fmt.Sprintf("%%%s%%", term))
	if err != nil {
		logger.WithError(err).Errorln("Can't read sites")
		return []models.Site{}
	}
	return items
}
func (pgr *PostgresRepository) GetSimpleSites() []models.Site {
	logger := pgr.Logger
	items := []models.Site{}
	err := pgr.sqlEngine.Select(&items, `SELECT 
			code,url,name,enabled
		FROM 
			site`)
	if err != nil {
		logger.WithError(err).Errorln("Can't read sites")
		return []models.Site{}
	}
	return items
}

func (pgr *PostgresRepository) GetSitesByOrg(orgID string) []models.Site {
	logger := pgr.Logger.WithField("orgId", orgID)
	items := []models.Site{}
	err := pgr.sqlEngine.Select(&items, `SELECT 
			code, url, name, group_name, createdDate, updatedDate, enabled 
		FROM  site
		INNER JOIN site_organisation so ON so.site_code = site.code WHERE so.org_id=$1`, orgID)
	if err != nil {
		logger.WithError(err).Errorln("Can't read sites")
		return []models.Site{}
	}
	return items
}

func (pgr *PostgresRepository) LinkSiteToOrg(orgID string, siteCodes []string) error {
	logger := pgr.Logger.WithFields(logrus.Fields{"orgId": orgID})
	now := time.Now()
	tx, err := pgr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(` INSERT INTO site_organisation (site_code, org_id, created) VALUES ($1,$2,$3) `)
	if err == nil {
		defer stmt.Close()
		for _, siteCode := range siteCodes {
			_, err = stmt.Exec(siteCode, orgID, now)
			if err != nil {
				logger.WithError(err).Errorln(err)
				return err
			}
		}
		tx.Commit()
	} else {
		logger.WithError(err).Error("can't create a statement to link org to site")
	}
	return nil
}
func (pgr *PostgresRepository) UnLinkSiteToOrg(orgID string, siteCodes []string) error {
	logger := pgr.Logger.WithFields(logrus.Fields{"orgId": orgID})
	tx, err := pgr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`DELETE FROM site_organisation WHERE site_code = $1 AND org_id=$2`)
	if err == nil {
		defer stmt.Close()
		for _, siteCode := range siteCodes {
			_, err = stmt.Exec(siteCode, orgID)
			if err != nil {
				logger.WithError(err).Errorln(err)
				return err
			}
		}
		tx.Commit()
	} else {
		logger.WithError(err).Error("can't create a statement to unlink org to site")
	}
	return nil
}

func (pgr *PostgresRepository) GetSite(code string) (site *models.Site) {
	logger := pgr.Logger
	rows, err := pgr.db.Query("SELECT  code,url,name,config,group_name,createdDate,updatedDate,enabled  FROM site WHERE code = $1", code)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var data []byte
			var dbSite models.Site
			err = rows.Scan(&dbSite.Code, &dbSite.URL, &dbSite.Name, &data, &dbSite.GroupName, &dbSite.CreatedDate, &dbSite.UpdatedDate, &dbSite.Enabled)
			if err == nil {
				site = &dbSite
			} else {
				logger.WithError(err).Errorf("Can't read site %s", code)
			}
		}
	} else {
		pgr.Logger.WithError(err).Errorf("can't get site %s", code)
	}
	return
}

func (pgr *PostgresRepository) GetSiteWithConfig(code string) (site *models.Site) {
	logger := pgr.Logger
	rows, err := pgr.db.Query(`SELECT  
		code,url,name,group_name,createdDate,updatedDate,enabled, 
		COALESCE(site_config, '{}') as site_config, COALESCE(site_config_draft, '{}')  as site_config_draft
	FROM site 
	WHERE 
		code = $1`, code)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var portalConfigData, portalConfigDraftData []byte
			var portalConfig models.PortalConfig
			var portalConfigDraft models.PortalConfig
			dbSite := models.Site{}
			err = rows.Scan(&dbSite.Code, &dbSite.URL, &dbSite.Name, &dbSite.GroupName, &dbSite.CreatedDate, &dbSite.UpdatedDate, &dbSite.Enabled, &portalConfigData, &portalConfigDraftData)
			if err == nil {

				parseErr := json.Unmarshal(portalConfigData, &portalConfig)
				if parseErr != nil {
					logger.WithError(parseErr).Errorf("Can't read portal config for site %s", code)
				} else {
					dbSite.PortalConfig = &portalConfig
				}

				parseErr = json.Unmarshal(portalConfigDraftData, &portalConfigDraft)
				if parseErr != nil {
					logger.WithError(parseErr).Errorf("Can't read portal config draft for site %s", code)
				} else {
					dbSite.PortalConfigDraft = &portalConfigDraft
				}

				site = &dbSite

			} else {
				logger.WithError(err).Errorf("Can't read site %s", code)
			}
		}
	} else {
		pgr.Logger.WithError(err).Errorf("can't get site %s", code)
	}
	return
}

func (pgr *PostgresRepository) GetSiteConfigs(code string) []models.SiteConfigDetail {
	var items []models.SiteConfigDetail
	logger := pgr.Logger
	rows, err := pgr.db.Query("SELECT  id,site_code,created_date FROM site_config WHERE site_code = $1", code)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var siteConfig models.SiteConfigDetail
			err = rows.Scan(&siteConfig.ID, &siteConfig.SiteCode, &siteConfig.CreatedDate)
			if err != nil {
				logger.WithError(err).Errorln("Can't read sites ")
			} else {
				items = append(items, siteConfig)
			}
		}
	} else {
		pgr.Logger.WithError(err).Errorln("can't GetSites", err)
	}
	return items
}

func (pgr *PostgresRepository) GetAreaListingCount(areaID int) int64 {
	rows, err := pgr.db.Query(sqlQueryAreaListingCount, areaID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var totalListing sql.NullInt64
			err = rows.Scan(&totalListing)
			if err == nil {
				return totalListing.Int64
			}
		}
	}
	if err != nil {
		pgr.Logger.WithError(err).Errorf("can't get total of listing for area %d", areaID)
	}
	return -1
}

func (pgr *PostgresRepository) GetAreas() []*models.Area {
	logger := pgr.Logger
	var items []*models.Area
	rows, err := pgr.db.Query(`SELECT name,ogc_fid as id,ST_AsGeoJSON(wkb_geometry) as geoJSON, 'borough' AS type ,'UK' AS country ,'greater_london' AS region from city_boundary 
	UNION
	SELECT location.name,location.id,ST_AsGeoJSON(location.geo)  as geoJSON,'area' AS type ,'UK' AS country,'greater_london' AS region  FROM location`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var area models.Area
			err := rows.Scan(&area.Name, &area.ID, &area.GeoJSON, &area.AreaType, &area.Country, &area.Region)
			if err == nil {
				items = append(items, &area)
			} else {
				logger.WithError(err).Errorf("Can't read areas")
				return items
			}
		}
	} else {
		pgr.Logger.WithError(err).Errorln("can't GetAreas", err)
	}
	return items
}

func (pgr *PostgresRepository) SaveSiteConfig(siteCode string, schedules []models.ScheduleForm) (err error) {
	var tr *sql.Tx
	tr, err = pgr.db.Begin()
	if err == nil {
		var stmt *sql.Stmt
		now := time.Now()
		stmt, err = tr.Prepare(`INSERT INTO scrape_area (site, geomid, enabled, createddate, updateddate, max_retry, frequency) 
				(SELECT $1, ogc_fid,$2,$3,$4,$5,$6 FROM city_boundary WHERE name = $7)
				ON CONFLICT (site, geomid) DO UPDATE SET max_retry = EXCLUDED.max_retry, enabled = EXCLUDED.enabled, frequency = EXCLUDED.frequency, updateddate= $4`)
		if err == nil {
			defer stmt.Close()
			for i := 0; i < len(schedules); i++ {
				schedule := schedules[i]
				enabled := 0

				if schedule.Enabled {
					enabled = 1
				}

				_, err = stmt.Exec(siteCode, enabled, now, now, schedule.MaxRetry, schedule.Frequency, schedule.AreaCode)
				if err != nil {
					tr.Rollback()
					pgr.Logger.WithError(err).Errorln(err)
					return
				}
			}
		} else {
			pgr.Logger.WithError(err).Errorln("can't create a statement", err)
		}
		tr.Commit()
	} else {
		pgr.Logger.WithError(err).Errorln("can't create listings", err)
	}
	return
}

func (pgr *PostgresRepository) GetScraperDetailsBySite(siteCode string) []models.ScrapingDetails {
	scrapers := []models.ScrapingDetails{}
	rows, err := pgr.db.Query(`SELECT
	sca.session_id, COALESCE(sca.lockedby,''), sca.site, cb.name AS areaCode, cb.ogc_fid AS areaId,
			sca.createddate, sca.last_success_date, sca.frequency, sca.enabled, COALESCE(sca.lockedby,''), sca.max_retry, sca.updateddate
		FROM
			scrape_area sca
			INNER JOIN city_boundary cb ON cb.ogc_fid = sca.geomid
		WHERE
			sca.site = $1
		ORDER BY
			sca.last_success_date desc`, siteCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var scraper models.ScrapingDetails
			var sessionID sql.NullString
			if err = rows.Scan(&sessionID, &scraper.UserID, &scraper.Site,
				&scraper.AreaCode, &scraper.AreaID, &scraper.CreatedDate, &scraper.LastSuccessDate, &scraper.Frequency,
				&scraper.Enabled, &scraper.LockedBy, &scraper.MaxRetry, &scraper.UpdatedDate); err == nil {
				scraper.SessionID = sessionID.String
				scrapers = append(scrapers, scraper)
			} else {
				break
			}
		}
	}

	if err != nil {
		pgr.Logger.WithError(err).Error("Can't get scrapers details")
	}
	return scrapers
}

func (pgr *PostgresRepository) handleAreaClusterResultSet(rows *sql.Rows, handler ClusterHandler) {
	defer rows.Close()
	for rows.Next() {

		ids := ""
		address := ""
		cid := ""
		location := "{}"
		images := ""

		var rawids sql.NullString
		var rawaddress sql.NullString
		var rawcid sql.NullString
		var floors int32
		var floorArea int32
		var beds int32
		var ppw int32
		var baths int32
		var floorPlan bool
		var rawLocation sql.NullString
		var rawImages sql.NullString
		var rawFloorPlan sql.NullInt64

		var rawIsUnderOffer sql.NullBool
		var rawIsLet sql.NullBool
		var rawIsDelisted sql.NullBool

		err := rows.Scan(&rawids, &rawaddress, &rawcid,
			&floors, &beds, &ppw, &floorArea, &baths, &rawLocation, &rawImages, &rawFloorPlan,
			&rawIsDelisted, &rawIsLet, &rawIsUnderOffer)

		if err == nil {
			if rawids.Valid {
				ids = rawids.String
			}

			if rawaddress.Valid {
				address = rawaddress.String
			}

			if rawcid.Valid {
				cid = rawcid.String
			}

			if rawLocation.Valid {
				location = rawLocation.String
			}

			if rawImages.Valid {
				images = rawImages.String
			}

			if rawFloorPlan.Valid {
				floorPlan = rawFloorPlan.Int64 > 0
			}

			c := models.RawAreaCluster{
				Ids:        ids,
				Address:    address,
				Cid:        cid,
				Floors:     floors,
				Beds:       beds,
				Ppw:        ppw,
				Baths:      baths,
				Location:   location,
				Images:     images,
				FloorArea:  floorArea,
				FloorPlan:  floorPlan,
				Delisted:   rawIsDelisted.Bool,
				UnderOffer: rawIsUnderOffer.Bool,
				Let:        rawIsLet.Bool,
			}

			if !handler(&c, false) {
				break
			}
		} else {
			pgr.Logger.WithError(err).Error("can't get properties")
		}
	}
	handler(nil, true)
}

func (pgr *PostgresRepository) GetAreaCluster(areaId int, handler ClusterHandler) {
	pgr.Logger.Infof("GetAreaCluster %d", areaId)
	logger := pgr.Logger
	rows, err := pgr.db.Query(sqlQueryCluster, areaId)
	if err == nil {
		pgr.handleAreaClusterResultSet(rows, handler)
	} else {
		logger.WithError(err).Errorln("can't GetAreas clusters")
	}
}

func (pgr *PostgresRepository) GetListingCluster(id models.ListingKey, handler ClusterHandler) {
	logger := pgr.Logger
	areaIDRow := pgr.db.QueryRow(`SELECT cb.ogc_fid FROM listing l
	LEFT JOIN city_boundary cb on st_contains(cb.wkb_geometry,l.geom)  AND cb.ogc_fid <> 0 
	WHERE value ->> 'id' = $1 AND  value ->> 'site' = $2`, id.ID, id.Site)
	var areaID sql.NullInt64
	err := areaIDRow.Scan(&areaID)
	if err == nil {
		var rows *sql.Rows

		if areaID.Valid {
			rows, err = pgr.db.Query(fmt.Sprintf("SELECT * FROM ( %s ) as allQuery WHERE allQuery.ids like $2 ",
				sqlQueryCluster), areaID.Int64, fmt.Sprintf("%%%s_%s%%", id.Site, id.ID))
		} else {
			rows, err = pgr.db.Query(sqlListingSingleQuery, id.Site, id.ID)
		}

		if err == nil {
			pgr.handleAreaClusterResultSet(rows, handler)
		}
	}

	if err != nil {
		logger.WithError(err).Errorf("can't Get area clusters for listing %s for site %s", id.ID, id.Site)
	}

}

const sqlQueryCluster = `SELECT string_agg(myquery.site::text || '_'|| myquery.id,',') AS ids,max(address) ,
myquery.cid,
	myquery.floors,
	myquery.beds ,
	myquery.ppw ,
	myquery.floorArea ,
	max(myquery.baths) ,
	max(location) as location,
	max(images)::json ->> 0 as images,
	length((max(floorPlans)::json  ->> 0)) as floorPlans,
  bool_or(myquery.isDelisted) as isDelisted,
  bool_or(myquery.isLet) as isLet ,
  bool_or(myquery.isUnderOffer) as isUnderOffer
FROM  (
SELECT
	created_date,
	updated_date,
	ST_ClusterDBSCAN(l.geom, eps := 0.00003, minpoints := 1) over () AS cid,
	delisted,
	value ->> 'id' AS id ,
	value ->> 'eventType' AS eventType ,
	value ->> 'site' AS site ,
	value ->> 'address_text' AS address ,
	value ->> 'floors'  AS floors,
	value ->> 'beds' AS beds ,
	value -> 'price'->> 'amount_week' AS ppw,
	value -> 'price'->> 'amount' AS ppm,
	value ->> 'baths' AS baths,
	value ->> 'reception' AS reception,
	value -> 'floorArea' ->> 'value' AS floorArea,
  (value ->> 'underOffer' )::bool AS isUnderOffer,
  (value ->> 'let' )::bool  AS isLet,
  (value ->> 'delisted' )::bool  AS isDelisted,
	'[' || (value -> 'location' ->> 'lon') || ',' || (value -> 'location' ->> 'lat') || ']' AS location,
	value ->> 'images' AS images, value ->> 'floorPlans' AS floorPlans
from listing l
	INNER JOIN city_boundary cb on st_contains(cb.wkb_geometry,l.geom)  AND cb.ogc_fid <> 0 
where cb.ogc_fid = $1
) myquery
group by
	myquery.cid,
	myquery.floors,
	myquery.beds ,
	myquery.ppw ,
	myquery.floorArea
order by cid,floors,beds,ppw,floorArea`

const sqlListingSingleQuery = `SELECT string_agg(myquery.site::text || '_'|| myquery.id,',') AS ids,max(address) ,
myquery.cid,
	myquery.floors,
	myquery.beds ,
	myquery.ppw ,
	myquery.floorArea ,
	max(myquery.baths) ,
	max(location) as location,
	max(images)::json ->> 0 as images,
	length((max(floorPlans)::json  ->> 0)) as floorPlans,
  bool_or(myquery.isDelisted) as isDelisted,
  bool_or(myquery.isLet) as isLet ,
  bool_or(myquery.isUnderOffer) as isUnderOffer
FROM  (
SELECT
	created_date,
	updated_date,
	1 as cid,
	delisted,
	value ->> 'id' AS id ,
	value ->> 'eventType' AS eventType ,
	value ->> 'site' AS site ,
	value ->> 'address_text' AS address ,
	value ->> 'floors'  AS floors,
	value ->> 'beds' AS beds ,
	value -> 'price'->> 'amount_week' AS ppw,
	value -> 'price'->> 'amount' AS ppm,
	value ->> 'baths' AS baths,
	value ->> 'reception' AS reception,
	value -> 'floorArea' ->> 'value' AS floorArea,
  (value ->> 'underOffer' )::bool AS isUnderOffer,
  (value ->> 'let' )::bool  AS isLet,
  (value ->> 'delisted' )::bool  AS isDelisted,
	'[' || coalesce(value -> 'location' ->> 'lon','0') || ',' ||  coalesce(value -> 'location' ->> 'lat','0') || ']' AS location,
	value ->> 'images' AS images, value ->> 'floorPlans' AS floorPlans
from listing l WHERE  l.value ->> 'site' = $1 AND  l.value ->> 'id' = $2
) myquery
group by
	myquery.cid,
	myquery.floors,
	myquery.beds ,
	myquery.ppw ,
	myquery.floorArea
order by cid,floors,beds,ppw,floorArea`

const sqlQueryAreaListingCount = `SELECT
count(id)
from listing l
INNER JOIN city_boundary cb on st_contains(cb.wkb_geometry,l.geom)  AND cb.ogc_fid <> 0 
where cb.ogc_fid = $1 and delisted = false`

const sqlQueryAreaListing = `SELECT
(value ->> 'site')::text || '_'|| (value ->> 'id')::text AS id,
value ->> 'site' AS site ,
value ->> 'address_text' AS address ,
value ->> 'floors'  AS floors,
value ->> 'beds' AS beds ,
value -> 'price'->> 'amount_week' AS ppw,
value ->> 'baths' AS baths,
value -> 'floorArea' ->> 'value' AS floorArea,
value -> 'location' ->> 'lon' as longitude,
value -> 'location' ->> 'lat' as latitude,
(value ->> 'underOffer' )::bool AS isUnderOffer,
(value ->> 'let' )::bool  AS isLet,
(value ->> 'delisted' )::bool  AS isDelisted,
(value ->> 'images')::json ->> 0 as images,
length(((value ->> 'floorPlans')::json  ->> 0)) as floorPlans
from listing l
INNER JOIN city_boundary cb on st_contains(cb.wkb_geometry,l.geom)  AND cb.ogc_fid <> 0 
where cb.ogc_fid = $1 and delisted = false`

const sqlQueryAllListing = `SELECT
	(value ->> 'site')::text || '_'|| (value ->> 'id')::text AS id,
	value ->> 'floors'  AS floors,
	value ->> 'beds' AS beds ,
	value -> 'price'->> 'amount_week' AS ppw,
	value ->> 'baths' AS baths,
	value -> 'floorArea' ->> 'value' AS floorArea,
	value -> 'location' ->> 'lon' as longitude,
	value -> 'location' ->> 'lat' as latitude,
	(value ->> 'underOffer' )::bool AS isUnderOffer,
	(value ->> 'let' )::bool  AS isLet,
	(value ->> 'delisted' )::bool  AS isDelisted,
	created_date
	FROM listing l
	WHERE value -> 'location' ->> 'lon'  IS NOT null AND  created_date > $1  AND delisted  = false order by created_date asc
`

type sqlValues struct {
	Args  [][]interface{}
	Error error
}
