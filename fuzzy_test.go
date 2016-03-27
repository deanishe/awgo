package workflow

import "testing"

var hostnames []string

// Data model
type Host struct {
	Hostname string
}

// Implements Fuzzy interface
type Hosts []Host

func (slice Hosts) Len() int {
	return len(slice)
}

func (slice Hosts) Less(i, j int) bool {
	return slice[i].Hostname < slice[j].Hostname
}

func (slice Hosts) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice Hosts) Keywords(i int) string {
	return slice[i].Hostname
}

// Test case
type hostTest struct {
	hosts    Hosts
	query    string
	expected string
}

// Generate test dataset
func makeHosts(hostnames []string) Hosts {
	hosts := make(Hosts, len(hostnames))
	for i, name := range hostnames {
		hosts[i] = Host{name}
	}
	return hosts
}

func TestFuzzy(t *testing.T) {
	simpleHostnames := []string{
		"www.example.com",
		"one.example.com",
		"two.example.com",
		"www.google.com",
		"www.amazon.de",
		// Contains "two"
		"www.two.co.uk",
	}
	initialHostnames := []string{
		// Initials "two"
		"test.work.org",
	}
	var allHostnames []string
	for _, s := range simpleHostnames {
		allHostnames = append(allHostnames, s)
	}
	for _, s := range initialHostnames {
		allHostnames = append(allHostnames, s)
	}

	simpleHosts := makeHosts(simpleHostnames)
	initialHosts := makeHosts(initialHostnames)
	allHosts := makeHosts(allHostnames)

	tests := []hostTest{
		// Search on prefix
		hostTest{simpleHosts, "two", "two.example.com"},
		hostTest{simpleHosts, "one", "one.example.com"},
		// Search on contains
		hostTest{simpleHosts, "ama", "www.amazon.de"},
		hostTest{simpleHosts, "ple", "one.example.com"},
		hostTest{simpleHosts, "two", "two.example.com"},
		// Fall back to normal sort here
		hostTest{simpleHosts, "example", "one.example.com"},
		// Search on initials
		hostTest{simpleHosts, "wec", "www.example.com"},
		hostTest{simpleHosts, "wad", "www.amazon.de"},
		hostTest{initialHosts, "two", "test.work.org"},
		// There is a conflicting prefix match
		hostTest{allHosts, "two", "test.work.org"},
	}
	for _, ht := range tests {
		SortFuzzy(ht.hosts, ht.query)
		if ht.hosts[0].Hostname != ht.expected {
			t.Fatalf("query=%v --> %v (actual) != %v (expected)", ht.query,
				ht.hosts[0].Hostname, ht.expected)
		}
	}
}
