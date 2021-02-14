CREATE TABLE project  (
    id varchar(35) NOT NULL,
    name varchar(50) ,
    description varchar(200) ,
    active boolean DEFAULT true,
    created timestamp ,
    updated timestamp ,
    CONSTRAINT "project_pkey" PRIMARY KEY (id)
);

CREATE TABLE application  (
    id varchar(35) NOT NULL,
    name varchar(50) ,
    description varchar(200) ,
    created timestamp ,
    updated timestamp ,
    CONSTRAINT "application_pkey" PRIMARY KEY (id)
);

CREATE TABLE webhook_type (
    id varchar(35) NOT NULL,
    code varchar(50) ,
    name varchar(50) ,
    description varchar(200) ,
    config text  NOT NULL,
    active boolean DEFAULT true,
    created timestamp ,
    updated timestamp ,
    CONSTRAINT "webhook_type_pkey" PRIMARY KEY (id)
);

CREATE TABLE webhook  (
    id varchar(35) NOT NULL,
    project_id varchar(35) ,
    name varchar(50) ,
    description varchar(200) ,
    active boolean DEFAULT true,
    created timestamp ,
    updated timestamp ,
    webhook_type_id varchar(35),
    CONSTRAINT "webhook_pkey" PRIMARY KEY (id),
    CONSTRAINT fk_wh_webhook_type FOREIGN KEY(webhook_type_id)  REFERENCES webhook_type(id),
    CONSTRAINT fk_wh_project FOREIGN KEY(project_id)  REFERENCES project(id)
);

CREATE TABLE webhook_config  (
    webhook_id varchar(35) NOT NULL,
    config text  NOT NULL,
    created timestamp ,
    updated timestamp ,
    CONSTRAINT "webhook_config_pkey" PRIMARY KEY (webhook_id),
    CONSTRAINT fk_wh_wh_config FOREIGN KEY(webhook_id)  REFERENCES webhook(id)
);

CREATE TABLE webhook_config_version  (
    webhook_id varchar(35) NOT NULL,
    version int  NOT NULL,
    config text  NOT NULL,
    created timestamp ,
    CONSTRAINT "webhook_config_version_pkey" PRIMARY KEY (webhook_id, version),
    CONSTRAINT fk_wh_wh_config FOREIGN KEY(webhook_id)  REFERENCES webhook(id)
);

CREATE TABLE webhook_handler_type (
    id varchar(35) NOT NULL,
    name varchar(50) ,
    description varchar(200) ,
    config text  NOT NULL,
    active boolean DEFAULT true,
    created timestamp ,
    updated timestamp ,
    CONSTRAINT "webhook_handler_type_pkey" PRIMARY KEY (id)
);

CREATE TABLE webhook_handler  (
    webhook_id varchar(35) NOT NULL,
    name varchar(50) ,
    description varchar(200) ,
    config text  NOT NULL,
    active boolean DEFAULT true,
    created timestamp ,
    updated timestamp ,
    webhook_handler_type_id varchar(35) NOT NULL,
    CONSTRAINT "webhook_handler_pkey" PRIMARY KEY (webhook_id),
    CONSTRAINT fk_wh_wh_handler FOREIGN KEY(webhook_id)  REFERENCES webhook(id),
    CONSTRAINT fk_wht_wh_handler FOREIGN KEY(webhook_handler_type_id)  REFERENCES webhook_handler_type(id)
);

CREATE TABLE webhook_handler_config  (
    webhook_handler_id varchar(35) NOT NULL,
    config text  NOT NULL,
    created timestamp ,
    updated timestamp ,
    CONSTRAINT "webhook_handler_config_pkey" PRIMARY KEY (webhook_handler_id),
    CONSTRAINT fk_wh_wh_config FOREIGN KEY(webhook_handler_id)  REFERENCES webhook_handler(webhook_id)
);

CREATE TABLE webhook_handler_config_version  (
    webhook_handler_id varchar(35) NOT NULL,
    version int  NOT NULL,
    config text  NOT NULL,
    created timestamp ,
    CONSTRAINT "webhook_handler_config_version_pkey" PRIMARY KEY (webhook_handler_id, version),
    CONSTRAINT fk_wh_wh_config_version FOREIGN KEY(webhook_handler_id)  REFERENCES webhook_handler(webhook_id)
);