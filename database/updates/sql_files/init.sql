#version:1
# This SQL file is always the latest set of data, intended for a full reset.
CREATE TABLE rolereacts (
    channelID varchar(128),
    messageID varchar(128),
    roles     json,

    PRIMARY KEY (channelID, messageID)
);

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