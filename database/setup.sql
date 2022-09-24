CREATE TABLE rolereacts (
    channelID varchar(128),
    messageID varchar(128),
    roles json,

    PRIMARY KEY (channelID, messageID)
);