
-- +migrate Up
create table product_model_aliases (
    id serial primary key,
    manufacturer text not null,
    model text not null,
    model_alias text not null,
    unique (manufacturer, model, model_alias)
);

insert into product_model_aliases (manufacturer, model, model_alias) values
    ('Arai','Chaser','Vector'),
    ('Arai','Quantum ST','Signet-Q'),
    ('Arai','Quantum ST Pro','Signet-Q Pro Tour'),
    ('Arai','QV Pro','Signet-X'),
    ('Arai','Rebel','Defiant'),
    ('Arai','RX-7 GP','Corsair-V'),
    ('Arai','RX-7V','Corsair-X'),
    ('Arai','Viper GT','Profile'),
    ('Caberg','Duke and Tourmax','Duke'),
    ('LS2','FF 396','FT2 FF396'),
    ('LS2','FF320 Stream','Stream'),
    ('LS2','FF323 Arrow','Arrow'),
    ('LS2','FF325 Strobe','Strobe'),
    ('LS2','FF350 Mono','Mono'),
    ('LS2','FF366 Cyber','Cyber'),
    ('LS2','FF369 Delta','Delta'),
    ('LS2','FF375 Shogun','Shogun'),
    ('LS2','FF397 Vector','Vector'),
    ('LS2','MX419','Ohm'),
    ('LS2','MX456','Light'),
    ('Sedici','PRIMO STRADA','Strada Primo'),
    ('Sedici','Primo Strada Carbon','Strada Carbon Primo'),
    ('Shoei','NXR','RF-1200'),
    ('Shoei','Ryd','RF-SR'),
    ('Shoei','X Spirit','X-Eleven'),
    ('Shoei','X Spirit lll','X-Fourteen'),
    ('Shoei','XR-1000','RF-1000'),
    ('Shoei','XR-1100','RF-1100');

create table product_manufacturer_aliases (
    id serial primary key,
    manufacturer text not null,
    manufacturer_alias text not null,
    unique(manufacturer, manufacturer_alias)
);

insert into product_manufacturer_aliases (manufacturer, manufacturer_alias) values ('Kido', 'Scorpion');

-- +migrate Down
drop table product_manufacturer_aliases;
drop table product_model_aliases;