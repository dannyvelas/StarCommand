validateConfig return value representations:
* there is a validation error (missing value, innappropriate value)
* there is an internal error
* there is no error, but here's the map

map, error
* there is a validation error: (map, tableError)
* there is an internal error: (nil, internalError)
* there is no error but here's the map (map, nil)

|                        | SSH | YAML | ENV | BW |
|------------------------|-----|------|-----|----|
| INPUT hostName         | ✓   | ✓    |     |    |
| INPUT struct           | ✓   | ✓    | ✓   | ✓  |
| OUTPUT filledStruct    | ✓   | ✓    | ✓   | ✓  |
| OUTPUT validationTable | ✓   |      |     |    |
