package grinchtext

import "fmt"

func Health(percent float64) string {
	return fmt.Sprintf("👽 Grinch · %.2f/100%% HP", percent)
}
