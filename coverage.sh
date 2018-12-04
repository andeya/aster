#!/bin/bash

# Code coverage generation
COVERAGE_DIR="${COVERAGE_DIR:-coverage}"
# Specify a package
PKG_LIST=$(go list ./... | grep -v /vendor/)
if [ ! -z "$1" ]; then
    PKG_LIST=$(go list ./$1)
fi

# Create the coverage files directory
mkdir -p "$COVERAGE_DIR";

# Create a coverage file for each package
for package in ${PKG_LIST}; do
    go test -covermode=count -coverprofile "${COVERAGE_DIR}/${package##*/}.cov" "$package" ;
done ;

# Merge the coverage profile files
echo 'mode: count' > coverage.cov ;
tail -q -n +2 "${COVERAGE_DIR}"/*.cov >> coverage.cov ;

echo "start gen coverage."
# Display the global code coverage
go tool cover -func=coverage.cov ;

# generate HTML report
go tool cover -html=coverage.cov -o coverage.html ;

# Remove the coverage files directory
rm -rf "$COVERAGE_DIR";
