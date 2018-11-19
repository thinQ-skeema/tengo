package tengo

import (
	"testing"
)

func TestParseVendor(t *testing.T) {
	cases := map[string]Vendor{
		"MySQL Community Server (GPL)":                           VendorMySQL,
		"some random text MYSQL some random text":                VendorMySQL,
		"Percona Server (GPL), Release 84.0, Revision 47234b3":   VendorPercona,
		"Percona Server (GPL), Release '22', Revision 'f62d93c'": VendorPercona,
		"mariadb.org binary distribution":                        VendorMariaDB,
		"Source distribution":                                    VendorUnknown,
	}
	for input, expected := range cases {
		actual := ParseVendor(input)
		if actual != expected {
			t.Errorf("Expected ParseVendor(\"%s\") to return %s, instead found %s", input, expected, actual)
		}
	}
}

func TestParseVersion(t *testing.T) {
	cases := map[string][3]int{
		"5.6.40":                               {5, 6, 40},
		"5.7.22":                               {5, 7, 22},
		"5.6.40-84.0":                          {5, 6, 40},
		"5.7.22-22":                            {5, 7, 22},
		"10.1.34-MariaDB-1~jessie":             {10, 1, 34},
		"10.2.16-MariaDB-10.2.16+maria~jessie": {10, 2, 16},
		"10.3.7-MariaDB-1:10.3.7+maria~jessie": {10, 3, 7},
		"invalid":                 {0, 0, 0},
		"5":                       {0, 0, 0},
		"5.6.invalid":             {0, 0, 0},
		"5.7.9300000000000000000": {0, 0, 0},
	}
	for input, expected := range cases {
		actual := ParseVersion(input)
		if actual != expected {
			t.Errorf("Expected ParseVersion(\"%s\") to return %v, instead found %v", input, expected, actual)
		}
	}
}

func TestNewFlavor(t *testing.T) {
	type testcase struct {
		base            string
		versionParts    []int
		expected        Flavor
		expectedString  string
		expectSupported bool
	}
	cases := []testcase{
		{"mysql", []int{5, 6, 40}, FlavorMySQL56, "mysql:5.6", true},
		{"mysql:5.7", []int{}, FlavorMySQL57, "mysql:5.7", true},
		{"mysql:5.5.49", []int{}, FlavorMySQL55, "mysql:5.5", true},
		{"mysql", []int{8, 0, 11}, FlavorMySQL80, "mysql:8.0", true},
		{"mysql:8", []int{}, FlavorMySQL80, "mysql:8.0", true},
		{"mysql", []int{8, 1, 2}, Flavor{VendorMySQL, 8, 1}, "mysql:8.1", false},
		{"percona", []int{5, 6}, FlavorPercona56, "percona:5.6", true},
		{"percona:5.7", []int{}, FlavorPercona57, "percona:5.7", true},
		{"percona", []int{}, Flavor{VendorPercona, 0, 0}, "percona:0.0", false},
		{"percona", []int{8, 0, 12}, Flavor{VendorPercona, 8, 0}, "percona:8.0", false},
		{"mariadb", []int{10, 1, 10}, FlavorMariaDB101, "mariadb:10.1", true},
		{"mariadb:10.2", []int{}, FlavorMariaDB102, "mariadb:10.2", true},
		{"mariadb", []int{10, 3}, FlavorMariaDB103, "mariadb:10.3", true},
		{"mariadb", []int{10}, Flavor{VendorMariaDB, 10, 0}, "mariadb:10.0", false},
		{"webscalesql", []int{}, FlavorUnknown, "unknown:0.0", false},
		{"webscalesql", []int{5, 6}, Flavor{VendorUnknown, 5, 6}, "unknown:5.6", false},
	}
	for _, tc := range cases {
		fl := NewFlavor(tc.base, tc.versionParts...)
		if fl != tc.expected {
			t.Errorf("Unexpected return from NewFlavor: Expected %s, found %s", tc.expected, fl)
		} else if fl.String() != tc.expectedString {
			t.Errorf("Unexpected return from Flavor.String(): Expected %s, found %s", tc.expectedString, fl.String())
		} else if fl.Supported() != tc.expectSupported {
			t.Errorf("Unexpected return from Flavor.Supported(): Expected %t, found %t", tc.expectSupported, fl.Supported())
		}
	}
}

func TestFlavorVendorMinVersion(t *testing.T) {
	type testcase struct {
		receiver Flavor
		compare  Flavor
		expected bool
	}
	cases := []testcase{
		{FlavorMySQL56, FlavorMySQL56, true},
		{FlavorMySQL56, FlavorMySQL55, true},
		{FlavorMySQL56, FlavorMySQL57, false},
		{FlavorMySQL80, FlavorMySQL57, true},
		{FlavorMySQL56, FlavorPercona56, false},
		{FlavorMariaDB103, FlavorMySQL80, false},
	}
	for _, tc := range cases {
		actual := tc.receiver.VendorMinVersion(tc.compare.Vendor, tc.compare.Major, tc.compare.Minor)
		if actual != tc.expected {
			t.Errorf("Expected %s.VendorMinVersion(%s) to return %t, instead found %t", tc.receiver, tc.compare, tc.expected, actual)
		}
	}
}

func TestFlavorFractionalTimestamps(t *testing.T) {
	type testcase struct {
		receiver Flavor
		expected bool
	}
	cases := []testcase{
		{FlavorMySQL55, false},
		{FlavorMySQL56, true},
		{FlavorMySQL57, true},
		{FlavorMariaDB101, true},
		{NewFlavor("percona:5.5"), false},
		{FlavorPercona56, true},
		{FlavorUnknown, true},
	}
	for _, tc := range cases {
		actual := tc.receiver.FractionalTimestamps()
		if actual != tc.expected {
			t.Errorf("Expected %s.FractionalTimestamps() to return %t, instead found %t", tc.receiver, tc.expected, actual)
		}
	}
}

func TestFlavorHasDataDictionary(t *testing.T) {
	type testcase struct {
		receiver Flavor
		expected bool
	}
	cases := []testcase{
		{FlavorMySQL55, false},
		{FlavorMySQL57, false},
		{FlavorMySQL80, true},
		{FlavorMariaDB101, false},
		{NewFlavor("percona:8.0"), true},
		{FlavorPercona56, false},
		{FlavorUnknown, false},
	}
	for _, tc := range cases {
		actual := tc.receiver.HasDataDictionary()
		if actual != tc.expected {
			t.Errorf("Expected %s.HasDataDictionary() to return %t, instead found %t", tc.receiver, tc.expected, actual)
		}
	}
}

func TestFlavorDefaultUtf8mb4Collation(t *testing.T) {
	type testcase struct {
		receiver Flavor
		expected string
	}
	cases := []testcase{
		{FlavorMySQL55, "utf8mb4_general_ci"},
		{FlavorMySQL57, "utf8mb4_general_ci"},
		{FlavorMySQL80, "utf8mb4_0900_ai_ci"},
		{FlavorMariaDB101, "utf8mb4_general_ci"},
		{NewFlavor("percona:8.0"), "utf8mb4_0900_ai_ci"},
		{FlavorPercona56, "utf8mb4_general_ci"},
		{FlavorUnknown, "utf8mb4_general_ci"},
	}
	for _, tc := range cases {
		actual := tc.receiver.DefaultUtf8mb4Collation()
		if actual != tc.expected {
			t.Errorf("Expected %s.DefaultUtf8mb4Collation() to return %s, instead found %s", tc.receiver, tc.expected, actual)
		}
	}
}

func TestFlavorAlwaysShowTableCollation(t *testing.T) {
	type testcase struct {
		receiver Flavor
		charSet  string
		expected bool
	}
	cases := []testcase{
		{FlavorMySQL55, "utf8mb4", false},
		{FlavorMySQL57, "utf8", false},
		{FlavorMySQL80, "utf8mb4", true},
		{FlavorMySQL80, "latin1", false},
		{FlavorMariaDB101, "utf8mb4", false},
		{FlavorPercona56, "utf8", false},
		{NewFlavor("percona:8.0"), "utf8mb4", true},
		{NewFlavor("percona:8.0"), "utf8", false},
		{FlavorUnknown, "utf8mb4", false},
	}
	for _, tc := range cases {
		actual := tc.receiver.AlwaysShowTableCollation(tc.charSet)
		if actual != tc.expected {
			t.Errorf("Expected %s.AlwaysShowCollation(%s) to return %t, instead found %t", tc.receiver, tc.charSet, tc.expected, actual)
		}
	}

}

func TestFlavorHasInnoFileFormat(t *testing.T) {
	type testcase struct {
		receiver Flavor
		expected bool
	}
	cases := []testcase{
		{FlavorMySQL55, true},
		{FlavorMySQL57, true},
		{FlavorMySQL80, false},
		{FlavorMariaDB102, true},
		{FlavorMariaDB103, false},
		{FlavorPercona57, true},
		{NewFlavor("percona:8.0"), false},
		{FlavorUnknown, true},
	}
	for _, tc := range cases {
		actual := tc.receiver.HasInnoFileFormat()
		if actual != tc.expected {
			t.Errorf("Expected %s.HasInnoFileFormat() to return %t, instead found %t", tc.receiver, tc.expected, actual)
		}
	}
}

func (s TengoIntegrationSuite) TestFlavorHasInnoFileFormat(t *testing.T) {
	flavor := s.d.Flavor()
	db, err := s.d.Connect("", "")
	if err != nil {
		t.Fatalf("Unexpected error from Connect: %s", err)
	}
	var innoFileFormat string
	err = db.QueryRow("SELECT @@global.innodb_file_format").Scan(&innoFileFormat)
	expected := flavor.HasInnoFileFormat()
	actual := (err == nil)
	if expected != actual {
		t.Errorf("Flavor %s expected existence of innodb_file_format is %t, instead found %t", flavor, expected, actual)
	}
}

func TestInnoRowFormatReqs(t *testing.T) {
	type testcase struct {
		receiver           Flavor
		format             string
		expectFilePerTable bool
		expectBarracuda    bool
	}
	cases := []testcase{
		{FlavorMySQL55, "DYNAMIC", true, true},
		{FlavorMySQL56, "DYNAMIC", true, true},
		{FlavorMySQL57, "DYNAMIC", false, false},
		{FlavorMySQL57, "COMPRESSED", true, true},
		{FlavorMySQL57, "COMPACT", false, false},
		{FlavorMySQL80, "DYNAMIC", false, false},
		{FlavorMySQL80, "COMPRESSED", true, false},
		{FlavorPercona56, "COMPRESSED", true, true},
		{FlavorPercona57, "DYNAMIC", false, false},
		{FlavorMariaDB101, "DYNAMIC", false, false},
		{FlavorMariaDB101, "REDUNDANT", false, false},
		{FlavorMariaDB102, "DYNAMIC", false, true},
		{FlavorMariaDB102, "COMPRESSED", true, true},
		{FlavorMariaDB103, "DYNAMIC", false, false},
		{FlavorMariaDB103, "COMPRESSED", true, false},
		{NewFlavor("mariadb:5.5"), "DYNAMIC", true, true},
		{FlavorUnknown, "DYNAMIC", true, true},
		{FlavorUnknown, "COMPRESSED", true, true},
		{FlavorUnknown, "COMPACT", false, false},
	}
	for _, tc := range cases {
		actualFilePerTable, actualBarracuda := tc.receiver.InnoRowFormatReqs(tc.format)
		if actualFilePerTable != tc.expectFilePerTable || actualBarracuda != tc.expectBarracuda {
			t.Errorf("Expected %s.InnoRowFormatReqs(%s) to return %t,%t; instead found %t,%t", tc.receiver, tc.format, tc.expectFilePerTable, tc.expectBarracuda, actualFilePerTable, actualBarracuda)
		}
	}

	var didPanic bool
	defer func() {
		if recover() != nil {
			didPanic = true
		}
	}()
	FlavorMySQL80.InnoRowFormatReqs("SUPER-DUPER-FORMAT")
	if !didPanic {
		t.Errorf("Expected InnoRowFormatReqs to panic on invalid format, but it did not")
	}
}
