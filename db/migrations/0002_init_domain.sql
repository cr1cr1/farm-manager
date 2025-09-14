CREATE TABLE IF NOT EXISTS barns (
    barn_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    capacity INTEGER,
    environment_control TEXT,
    maintenance_schedule TEXT,
    location TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER
);

CREATE TABLE IF NOT EXISTS feed_types (
    feed_type_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    nutritional_info TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER
);

CREATE TABLE IF NOT EXISTS staff (
    staff_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    role TEXT,
    schedule TEXT,
    contact_info TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER
);

CREATE TABLE IF NOT EXISTS flocks (
    flock_id INTEGER PRIMARY KEY AUTOINCREMENT,
    breed TEXT NOT NULL,
    hatch_date DATE,
    number_of_birds INTEGER,
    current_age INTEGER,
    barn_id INTEGER,
    health_status TEXT,
    feed_type_id INTEGER,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (barn_id) REFERENCES barns(barn_id),
    FOREIGN KEY (feed_type_id) REFERENCES feed_types(feed_type_id)
);

CREATE INDEX IF NOT EXISTS idx_flock_barn ON flocks(barn_id);
CREATE INDEX IF NOT EXISTS idx_flock_feedtype ON flocks(feed_type_id);

CREATE TABLE IF NOT EXISTS feeding_records (
    feeding_record_id INTEGER PRIMARY KEY AUTOINCREMENT,
    flock_id INTEGER,
    feed_type_id INTEGER,
    amount_given REAL,
    date_time DATETIME,
    staff_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (flock_id) REFERENCES flocks(flock_id),
    FOREIGN KEY (feed_type_id) REFERENCES feed_types(feed_type_id),
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id)
);

CREATE INDEX IF NOT EXISTS idx_feedingrecord_flock ON feeding_records(flock_id);
CREATE INDEX IF NOT EXISTS idx_feedingrecord_feedtype ON feeding_records(feed_type_id);
CREATE INDEX IF NOT EXISTS idx_feedingrecord_staff ON feeding_records(staff_id);

CREATE TABLE IF NOT EXISTS health_checks (
    health_check_id INTEGER PRIMARY KEY AUTOINCREMENT,
    flock_id INTEGER,
    check_date DATETIME,
    health_status TEXT,
    vaccinations_given TEXT,
    treatments_administered TEXT,
    notes TEXT,
    staff_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (flock_id) REFERENCES flocks(flock_id),
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id)
);

CREATE INDEX IF NOT EXISTS idx_healthcheck_flock ON health_checks(flock_id);
CREATE INDEX IF NOT EXISTS idx_healthcheck_staff ON health_checks(staff_id);

CREATE TABLE IF NOT EXISTS mortality_records (
    mortality_record_id INTEGER PRIMARY KEY AUTOINCREMENT,
    flock_id INTEGER,
    date DATETIME,
    number_dead INTEGER,
    cause_of_death TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (flock_id) REFERENCES flocks(flock_id)
);

CREATE INDEX IF NOT EXISTS idx_mortalityrecord_flock ON mortality_records(flock_id);

CREATE TABLE IF NOT EXISTS production_batches (
    batch_id INTEGER PRIMARY KEY AUTOINCREMENT,
    flock_id INTEGER,
    date_ready DATETIME,
    number_in_batch INTEGER,
    weight_estimate REAL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (flock_id) REFERENCES flocks(flock_id)
);

CREATE INDEX IF NOT EXISTS idx_productionbatch_flock ON production_batches(flock_id);

CREATE TABLE IF NOT EXISTS slaughter_records (
    slaughter_id INTEGER PRIMARY KEY AUTOINCREMENT,
    batch_id INTEGER,
    date DATETIME,
    number_slaughtered INTEGER,
    meat_yield REAL,
    waste REAL,
    staff_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (batch_id) REFERENCES production_batches(batch_id),
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id)
);

CREATE INDEX IF NOT EXISTS idx_slaughterrecord_batch ON slaughter_records(batch_id);
CREATE INDEX IF NOT EXISTS idx_slaughterrecord_staff ON slaughter_records(staff_id);

CREATE TABLE IF NOT EXISTS inventory_items (
    inventory_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT,
    quantity REAL,
    unit TEXT,
    expiration_date DATE,
    supplier_info TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER
);

CREATE TABLE IF NOT EXISTS customers (
    customer_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    contact_info TEXT,
    delivery_address TEXT,
    customer_type TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER
);

CREATE TABLE IF NOT EXISTS orders (
    order_id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER,
    order_date DATETIME,
    delivery_date DATETIME,
    total_amount REAL,
    status TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id)
);

CREATE INDEX IF NOT EXISTS idx_order_customer ON orders(customer_id);

CREATE TABLE IF NOT EXISTS order_items (
    order_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER,
    product_description TEXT,
    quantity REAL,
    unit_price REAL,
    total_price REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    created_by INTEGER,
    updated_by INTEGER,
    FOREIGN KEY (order_id) REFERENCES orders(order_id)
);

CREATE INDEX IF NOT EXISTS idx_orderitem_order ON order_items(order_id);
