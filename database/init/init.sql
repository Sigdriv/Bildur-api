-- ============================================================
--  Tabell: images
-- ============================================================
CREATE TABLE "images" (
    "id"           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"         TEXT NOT NULL,           -- Filnavn fra opplasting
    "mimeType"     TEXT NOT NULL,           -- f.eks. image/jpeg
    "extension"    TEXT NOT NULL,           -- f.eks. jpg, png
    "bytes"        BIGINT NOT NULL,         -- Filstørrelse
    "storagePath"  TEXT NOT NULL,           -- Hvor fila ligger (disk, S3, osv.)
    "width"        INT,                     -- Bildebredde i piksler
    "height"       INT,                     -- Bildehøyde i piksler
    "createdAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
--  Tabell: imagePreviews
-- ============================================================
CREATE TABLE "imagePreviews" (
    "id"           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "imageId"      UUID NOT NULL REFERENCES "images"("id") ON DELETE CASCADE,
    "variantName"  TEXT NOT NULL,                 -- f.eks. 'thumb', 'medium'
    "storagePath"  TEXT NOT NULL,                 -- hvor preview-fila ligger
    "width"        INT,
    "height"       INT,
    "createdAt"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE ("imageId", "variantName")
);

-- ============================================================
--  Tabell: filters
-- ============================================================
CREATE TABLE "filters" (
    "id"          SERIAL PRIMARY KEY,
    "name"        TEXT NOT NULL UNIQUE,
    "description" TEXT
);

INSERT INTO "filters" ("name", "description") VALUES
    ('grayscale', 'Konverter bildet til svart-kvitt')
ON CONFLICT ("name") DO NOTHING;

-- ============================================================
--  Tabell: editedImages
-- ============================================================
CREATE TABLE "editedImages" (
    "id"              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "originalImageId" UUID NOT NULL REFERENCES "images"("id") ON DELETE CASCADE,
    "filterId"         INT REFERENCES "filters"("id"),
    "storagePath"     TEXT NOT NULL,
    "width"           INT,
    "height"          INT,
    "createdAt"       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
--  Tabell: collages
-- ============================================================
CREATE TABLE "collages" (
    "id"          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"        TEXT,
    "rows"        INT NOT NULL CHECK ("rows" > 0),
    "cols"        INT NOT NULL CHECK ("cols" > 0),
    "storagePath" TEXT NOT NULL,              -- hvor ferdig kollasj-bilde ligger
    "width"       INT,
    "height"      INT,
    "createdAt"   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
--  Tabell: collageImages
-- ============================================================
CREATE TABLE "collageImages" (
    "collageId"  UUID NOT NULL REFERENCES "collages"("id") ON DELETE CASCADE,
    "imageId"    UUID NOT NULL REFERENCES "images"("id") ON DELETE RESTRICT,
    "rowIndex"   INT NOT NULL CHECK ("rowIndex" >= 0),
    "colIndex"   INT NOT NULL CHECK ("colIndex" >= 0),
    PRIMARY KEY ("collageId", "rowIndex", "colIndex")
);
