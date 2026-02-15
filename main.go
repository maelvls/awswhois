package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"text/tabwriter"
)

const awsIPRangesURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"

type AWSIPRanges struct {
	SyncToken  string       `json:"syncToken"`
	CreateDate string       `json:"createDate"`
	Prefixes   []IPPrefix   `json:"prefixes"`
	IPv6Prefixes []IPv6Prefix `json:"ipv6_prefixes"`
}

type IPPrefix struct {
	IPPrefix           string `json:"ip_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

type IPv6Prefix struct {
	IPv6Prefix         string `json:"ipv6_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <ip-or-hostname>\n", os.Args[0])
		os.Exit(1)
	}

	input := os.Args[1]

	// Fetch AWS IP ranges
	ranges, err := fetchAWSIPRanges()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching AWS IP ranges: %v\n", err)
		os.Exit(1)
	}

	// Resolve input to IPs
	ips, err := resolveToIPs(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving %s: %v\n", input, err)
		os.Exit(1)
	}

	if len(ips) == 0 {
		fmt.Fprintf(os.Stderr, "No IP addresses found for %s\n", input)
		os.Exit(1)
	}

	// Check each IP against AWS ranges and collect results
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "IP\tPREFIX\tREGION\tSERVICE\tBORDER GROUP")

	found := false
	for _, ip := range ips {
		matches := findAWSMatches(ip, ranges)
		if len(matches) > 0 {
			found = true
			// Group matches by IP + Prefix + Region + NetworkBorderGroup
			grouped := groupMatches(matches)
			for _, group := range grouped {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					ip.String(),
					group.Prefix,
					group.Region,
					group.Services,
					group.NetworkBorderGroup)
			}
		} else {
			fmt.Fprintf(w, "%s\t-\t-\t-\t-\n", ip.String())
		}
	}
	w.Flush()

	if !found {
		os.Exit(1)
	}
}

func fetchAWSIPRanges() (*AWSIPRanges, error) {
	resp, err := http.Get(awsIPRangesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ranges AWSIPRanges
	if err := json.Unmarshal(body, &ranges); err != nil {
		return nil, err
	}

	return &ranges, nil
}

func resolveToIPs(input string) ([]net.IP, error) {
	// Try parsing as IP first
	if ip := net.ParseIP(input); ip != nil {
		return []net.IP{ip}, nil
	}

	// Otherwise, resolve as hostname
	ips, err := net.LookupIP(input)
	if err != nil {
		return nil, err
	}

	return ips, nil
}

type AWSMatch struct {
	Prefix             string
	Region             string
	Service            string
	NetworkBorderGroup string
}

type GroupedMatch struct {
	Prefix             string
	Region             string
	Services           string
	NetworkBorderGroup string
}

func findAWSMatches(ip net.IP, ranges *AWSIPRanges) []AWSMatch {
	var matches []AWSMatch

	// Check IPv4 ranges
	if ip.To4() != nil {
		for _, prefix := range ranges.Prefixes {
			_, ipNet, err := net.ParseCIDR(prefix.IPPrefix)
			if err != nil {
				continue
			}
			if ipNet.Contains(ip) {
				matches = append(matches, AWSMatch{
					Prefix:             prefix.IPPrefix,
					Region:             prefix.Region,
					Service:            prefix.Service,
					NetworkBorderGroup: prefix.NetworkBorderGroup,
				})
			}
		}
	} else {
		// Check IPv6 ranges
		for _, prefix := range ranges.IPv6Prefixes {
			_, ipNet, err := net.ParseCIDR(prefix.IPv6Prefix)
			if err != nil {
				continue
			}
			if ipNet.Contains(ip) {
				matches = append(matches, AWSMatch{
					Prefix:             prefix.IPv6Prefix,
					Region:             prefix.Region,
					Service:            prefix.Service,
					NetworkBorderGroup: prefix.NetworkBorderGroup,
				})
			}
		}
	}

	return matches
}

func groupMatches(matches []AWSMatch) []GroupedMatch {
	// Group by Prefix + Region + NetworkBorderGroup
	type groupKey struct {
		Prefix             string
		Region             string
		NetworkBorderGroup string
	}

	grouped := make(map[groupKey][]string)
	var keys []groupKey

	for _, match := range matches {
		key := groupKey{
			Prefix:             match.Prefix,
			Region:             match.Region,
			NetworkBorderGroup: match.NetworkBorderGroup,
		}

		if _, exists := grouped[key]; !exists {
			keys = append(keys, key)
			grouped[key] = []string{}
		}

		grouped[key] = append(grouped[key], match.Service)
	}

	// Convert to GroupedMatch slice
	var result []GroupedMatch
	for _, key := range keys {
		services := ""
		for i, svc := range grouped[key] {
			if i > 0 {
				services += ","
			}
			services += svc
		}

		result = append(result, GroupedMatch{
			Prefix:             key.Prefix,
			Region:             key.Region,
			Services:           services,
			NetworkBorderGroup: key.NetworkBorderGroup,
		})
	}

	return result
}
