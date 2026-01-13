FROM ghcr.io/osgeo/gdal:ubuntu-small-latest

# Install tools to download and extract data
RUN apt-get update && apt-get install -y curl unzip && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Define environment variables with defaults
ENV DATABASE_HOST=localhost
ENV DATABASE_PORT=5432
ENV DATABASE_DB=playground
ENV DATABASE_USER=postgres
ENV DATABASE_PASSWORD=postgres

# Download, unzip, and run ogr2ogr
# We use a shell script format to allow variable expansion
CMD export PGPASSWORD=${DATABASE_PASSWORD} && \
  curl -O https://naciscdn.org/naturalearth/10m/cultural/ne_10m_populated_places.zip && \
  unzip ne_10m_populated_places.zip && \
  ogr2ogr \
  -update \
  -append \
  -nln gis.populated_places \
  --config PG_USE_COPY YES \
  -nlt POINT \
  -makevalid \
  -where "scalerank IS NOT NULL AND labelrank IS NOT NULL AND labelrank >= 0 AND labelrank <= 10 AND featurecla <> '' AND name <> '' AND nameascii <> '' AND sov0name <> '' AND adm0name <> '' AND adm0_a3 <> '' AND ISO_A2 <> '-99' AND adm1name IS NOT NULL AND pop_max >= 0 AND min_zoom IS NOT NULL" \
  Pg:"dbname=${DATABASE_DB} host=${DATABASE_HOST} user=${DATABASE_USER} port=${DATABASE_PORT}" \
  ne_10m_populated_places.shp && \
  echo "Data load complete."
