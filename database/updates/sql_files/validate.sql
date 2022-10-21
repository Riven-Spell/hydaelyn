#version:0
# ^ For the sake of parsing, include version

SELECT channelID, messageID, roles FROM rolereacts;

SELECT guildID, guildEventID, eventData FROM events;

SELECT cfgKey, cfgVal FROM config;