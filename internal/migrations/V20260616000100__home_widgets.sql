-- Home screen carousel widgets (managed from admin panel).

CREATE TABLE IF NOT EXISTS home_widgets (
    id               BIGSERIAL PRIMARY KEY,
    sort_order       INTEGER NOT NULL DEFAULT 0,
    tag_text         VARCHAR(64) NOT NULL DEFAULT '',
    tag_bg           VARCHAR(64) NOT NULL DEFAULT 'rgba(60,200,100,0.85)',
    tag_color        VARCHAR(64) NOT NULL DEFAULT '#06291a',
    title            TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    background_style TEXT NOT NULL DEFAULT 'linear-gradient(135deg,#1a1030,#2a1840)',
    image_url        TEXT NOT NULL DEFAULT '',
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_home_widgets_active_sort
    ON home_widgets (is_active, sort_order, id);
