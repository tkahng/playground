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
CMD curl -O https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_0_countries.zip && \
  unzip ne_10m_admin_0_countries.zip && \
  ogr2ogr \
  -update \
  -append \
  -nln gis.countries \
  -nlt MULTIPOLYGON \
  -makevalid \
  -where "ISO_A2_EH <> '-99' AND ISO_A3_EH <> '-99'" \
  Pg:"dbname=${DB_NAME} host=${DB_HOST} user=${DB_USER} port=${DB_PORT}" \
  ne_10m_admin_0_countries.shp && \
  echo "Data load complete."
