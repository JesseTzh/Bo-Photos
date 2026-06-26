CREATE INDEX album_assets_album_sort_idx
    ON album_assets(album_id, sort, asset_id);

CREATE INDEX album_assets_asset_idx
    ON album_assets(asset_id);

CREATE INDEX asset_tags_tag_asset_idx
    ON asset_tags(tag_id, asset_id);

CREATE INDEX tags_parent_idx
    ON tags(parent_id);
