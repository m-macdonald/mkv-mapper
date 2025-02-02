package path

import (
    "os"
)

func Check(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true;
    }

    return false;
}
