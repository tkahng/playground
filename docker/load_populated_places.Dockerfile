FROM ghcr.io/osgeo/gdal:ubuntu-small-latest

# Install tools to download and extract data
RUN apt-get update && apt-get install -y curl unzip && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Define environment variables with defaults
ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_NAME=playground
ENV DB_USER=postgres
ENV PGPASSWORD=postgres

# Download, unzip, and run ogr2ogr
# We use a shell script format to allow variable expansion
CMD curl -O https://naciscdn.org/naturalearth/10m/cultural/ne_10m_populated_places.zip && \
  unzip ne_10m_populated_places.zip && \
  ogr2ogr \
  -update \
  -append \
  -nln gis.populated_places \
  -nlt POINT \
  -makevalid \
  -where "scalerank IS NOT NULL AND labelrank IS NOT NULL AND labelrank >= 0 AND labelrank <= 10 AND featurecla <> '' AND name <> '' AND nameascii <> '' AND sov0name <> '' AND adm0name <> '' AND adm0_a3 <> '' AND ISO_A2 <> '-99' AND adm1name IS NOT NULL AND pop_max >= 0 AND min_zoom IS NOT NULL" \
  Pg:"dbname=${DB_NAME} host=${DB_HOST} user=${DB_USER} port=${DB_PORT}" \
  ne_10m_populated_places.shp && \
  echo "Data load complete."
