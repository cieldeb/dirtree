// SPDX-License-Identifier: GPL-3.0-or-later
package main

// pluralize adds suffixes to element names depending on the case
func pluralize(base string, count int) string {
	if base == "subdirector" {
		if count == 1 {
			return "subdirectory"
		}
		return "subdirectories"
	}
	// For "file"
	if count == 1 {
		return base
	}
	return base + "s"
}
