package hook_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dwellersclub/contigus/hook"
	"github.com/dwellersclub/contigus/models"
	"github.com/stretchr/testify/assert"
)

func copyFolder(from, to string) error {
	return filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
        var relPath string = strings.Replace(path, from, "", 1)
        if relPath == "" {
            return nil
        }
        if info.IsDir() {
            return os.Mkdir(filepath.Join(to, relPath), 0755)
        } else {
            var data, err1 = ioutil.ReadFile(filepath.Join(from, relPath))
            if err1 != nil {
                return err1
            }
            return ioutil.WriteFile(filepath.Join(to, relPath), data, 0777)
        }
    })
}

func TestLoadingHooks(t *testing.T) {
	//set up folder
	dir, err := ioutil.TempDir("", "hooks-*-test")
	if err != nil {
		log.Fatal(err)
	}
	t.Log(dir)
	defer os.Remove(dir)

	err = copyFolder("fixtures/repository/hooks", dir)
	if err != nil {
		t.Fail()
	}

	totalConfigReloaded := 0 
	updater := func(configs []models.HookConfig) {
		totalConfigReloaded = totalConfigReloaded + len(configs)
	}

	repo := hook.NewFileBasedRepo(dir,updater, 1)
	if repo != nil {
		t.Fail()
	}
	defer  repo.Close()

	time.Sleep(2*time.Second)

	assert.Equal(t,4,totalConfigReloaded)

}

func TestReloadLoadingHooks(t *testing.T) {
	//set up folder
	dir, err := ioutil.TempDir("", "hooks-*-test")
	if err != nil {
		log.Fatal(err)
	}
	t.Log(dir)
	defer os.Remove(dir)

	err = copyFolder("fixtures/repository/hooks", dir)
	if err != nil {
		t.Fail()
	}

	expectedTotalConfig := 4
	totalConfigReloaded := 0 
	updater := func(configs []models.HookConfig) {
		totalConfigReloaded = totalConfigReloaded + len(configs)
	}

	repo := hook.NewFileBasedRepo(dir,updater, 1)
	if repo == nil {
		t.Fail()
		return 
	}
	defer  repo.Close()

	time.Sleep(2*time.Second)
	
	assert.Equal(t,expectedTotalConfig,totalConfigReloaded)

	additionalConfig := 4

	for i := 0; i < additionalConfig; i++ {
		file := fmt.Sprintf( "%s/loading/test-%d.json",dir,i)
		content := fmt.Sprintf( "{\"type\":\"github\" , \"id\":\"test_id_%d\" , \"urlContext\":\"/my_context\"}\n",i)
		ioutil.WriteFile(file, []byte(content), 0644)
	}

	time.Sleep(2*time.Second)

	assert.Equal(t,expectedTotalConfig+additionalConfig,totalConfigReloaded)
}



func TestReloadWithDeletion(t *testing.T) {
	//set up folder
	dir, err := ioutil.TempDir("", "hooks-*-test")
	if err != nil {
		log.Fatal(err)
	}
	t.Log(dir)
	defer os.Remove(dir)

	err = copyFolder("fixtures/repository/hooks", dir)
	if err != nil {
		t.Fail()
	}

	expectedTotalConfig := 4
	totalConfigReloaded := 0 
	deletedCount := 0 
	updater := func(configs []models.HookConfig) {
		totalConfigReloaded = totalConfigReloaded + len(configs)
		for _, config := range configs {
			if config.Deleted {
				deletedCount++
			}
		}
	}

	repo := hook.NewFileBasedRepo(dir,updater, 1)
	if repo == nil {
		t.Fail()
		return 
	}
	defer  repo.Close()

	time.Sleep(2*time.Second)
	
	assert.Equal(t,expectedTotalConfig,totalConfigReloaded)

	fileToDel := fmt.Sprintf("%s/loading/1.json",dir)
	delErr := os.Remove(fileToDel)
	if delErr != nil {
		assert.Fail(t, "can't delete [%s]", fileToDel)
		return 
	}

	time.Sleep(2*time.Second)
	assert.Equal(t,1,deletedCount)
	assert.Equal(t,expectedTotalConfig+1,totalConfigReloaded)

}
