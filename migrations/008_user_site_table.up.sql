CREATE TABLE user_sites (
    user_id INT NOT NULL,
    site_id INT NOT NULL,
    filters JSON NOT NULL
    PRIMARY KEY (user_id, site_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (site_id) REFERENCES site_scraping_config(id)
);