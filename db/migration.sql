CREATE TABLE IF NOT EXISTS vehicle_locations (
    vehicle_id  VARCHAR(50)  NOT NULL,
    latitude    NUMERIC(9, 6)  NOT NULL,
    longitude   NUMERIC(9, 6)  NOT NULL,
    "timestamp" BIGINT        NOT NULL,
    PRIMARY KEY (vehicle_id, "timestamp")
);
