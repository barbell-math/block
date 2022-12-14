package settings;

import (
    customerr "github.com/barbell-math/block/util/err"
)

var DataVersionMalformed,IsDataVersionMalformed=customerr.ErrorFactory(
    "Data version is malformed.",
);

var SettingsFileNotFound,IsSettingsFileNotFound=customerr.ErrorFactory(
    "A file specified in the settings config file could not be found.",
);
