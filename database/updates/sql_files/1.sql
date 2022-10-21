#version:1

CREATE TABLE events (
    guildID        varchar(128),
    guildEventID   varchar(128),
    eventData      json,

    PRIMARY KEY (guildID, guildEventID)
);

CREATE TABLE config ( # Currently, only used to handle upgrades. Could be used for more in the future.
    cfgKey varchar(128),
    cfgVal json,

    PRIMARY KEY (cfgKey)
);