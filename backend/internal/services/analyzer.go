package services

import (
	"dnsmesh/internal/models"
	"fmt"
	"regexp"
	"sort"
)

// RegionMap maps region codes to region names
var RegionMap = map[string]string{
	// 国际
	"hk": "香港", "us": "美国", "sg": "新加坡",
	"jp": "日本", "kr": "韩国", "de": "德国",
	"uk": "英国", "cn": "中国", "au": "澳大利亚",
	"ca": "加拿大", "fr": "法国", "in": "印度",
	"tw": "台湾", "th": "泰国", "id": "印尼",
	// 中国城市
	"bj": "北京", "sh": "上海", "gz": "广州", "sz": "深圳",
	"cd": "成都", "cq": "重庆", "wh": "武汉", "xa": "西安",
	"hz": "杭州", "nj": "南京", "tj": "天津", "qd": "青岛",
	"dl": "大连", "sy": "沈阳", "cs": "长沙", "zz": "郑州",
}

// serverPattern matches region-prefixed domains with optional numeric suffixes
// Examples: hk.example, hk-1.example, hk1.example
var serverPattern = regexp.MustCompile(`^([a-z]{2,3})(?:-?(\d+))?\.`)

// AnalyzeDNSRecords analyzes DNS records and suggests servers
func AnalyzeDNSRecords(records []DNSRecordSync) AnalysisResult {
	var suggestions []ServerSuggestion

	// Build maps for analysis
	cnameTargetMap := make(map[string][]string) // target -> []sources
	ipMap := make(map[string][]string)          // ip -> []domains
	domainMap := make(map[string]DNSRecordSync) // domain -> record

	// First pass: build maps
	for _, record := range records {
		domainMap[record.FullDomain] = record

		if record.RecordType == models.RecordTypeCNAME {
			// Normalize trailing dot for CNAME targets (FQDN format)
			target := record.TargetValue
			if len(target) > 0 && target[len(target)-1] == '.' {
				target = target[:len(target)-1]
			}
			cnameTargetMap[target] = append(
				cnameTargetMap[target],
				record.FullDomain,
			)
		} else if record.RecordType == models.RecordTypeA {
			ipMap[record.TargetValue] = append(
				ipMap[record.TargetValue],
				record.FullDomain,
			)
		}
	}

	// Track already suggested domains to avoid duplicates
	suggested := make(map[string]bool)

	// Priority 1: Pattern matching (region-number format)
	for _, record := range records {
		if record.RecordType != models.RecordTypeA {
			continue
		}

		// Remove wildcard prefix for pattern matching
		domain := record.FullDomain
		isWildcard := false
		if len(domain) > 2 && domain[:2] == "*." {
			domain = domain[2:]
			isWildcard = true
		}

		matches := serverPattern.FindStringSubmatch(domain)
		if len(matches) > 2 {
			// Skip wildcard domains as server candidates
			// Wildcard records should be grouped under their base server
			if isWildcard {
				continue
			}

			regionCode := matches[1]
			serverNum := matches[2]

			suggestion := ServerSuggestion{
				Domain:      record.FullDomain,
				IP:          record.TargetValue,
				MatchReason: "域名格式匹配（地域-数字）",
				Confidence:  "high",
				SuggestedName: func() string {
					if serverNum != "" {
						return regionCode + "-" + serverNum
					}
					return regionCode
				}(),
				SuggestedRegion: RegionMap[regionCode],
				ReferencedBy:    cnameTargetMap[record.FullDomain],
			}

			// Add same IP domains
			sameIPDomains := []string{}
			for _, d := range ipMap[record.TargetValue] {
				if d != record.FullDomain {
					sameIPDomains = append(sameIPDomains, d)
				}
			}
			suggestion.SameIPDomains = sameIPDomains

			if len(suggestion.ReferencedBy) > 0 {
				suggestion.MatchReason += fmt.Sprintf(" + %d 个 CNAME 引用", len(suggestion.ReferencedBy))
			}

			suggestions = append(suggestions, suggestion)
			suggested[record.FullDomain] = true
		}
	}

	// Priority 2: CNAME reference analysis (not already matched by pattern)
	for _, record := range records {
		if record.RecordType != models.RecordTypeA || suggested[record.FullDomain] {
			continue
		}

		referencedBy := cnameTargetMap[record.FullDomain]
		if len(referencedBy) >= 2 {
			suggestion := ServerSuggestion{
				Domain:       record.FullDomain,
				IP:           record.TargetValue,
				MatchReason:  fmt.Sprintf("被 %d 个域名 CNAME 引用", len(referencedBy)),
				Confidence:   "medium",
				ReferencedBy: referencedBy,
			}

			// Add same IP domains
			sameIPDomains := []string{}
			for _, d := range ipMap[record.TargetValue] {
				if d != record.FullDomain {
					sameIPDomains = append(sameIPDomains, d)
				}
			}
			suggestion.SameIPDomains = sameIPDomains

			suggestions = append(suggestions, suggestion)
			suggested[record.FullDomain] = true
		}
	}

	// Priority 3: IP aggregation (multiple domains pointing to same IP)
	for ip, domains := range ipMap {
		if len(domains) >= 3 {
			// Check if any domain already suggested
			alreadySuggested := false
			for _, d := range domains {
				if suggested[d] {
					alreadySuggested = true
					break
				}
			}

			if !alreadySuggested {
				suggestion := ServerSuggestion{
					IP:            ip,
					MatchReason:   fmt.Sprintf("%d 个域名使用同一 IP", len(domains)),
					Confidence:    "low",
					SameIPDomains: domains,
				}

				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return AnalysisResult{
		Records:           records,
		ServerSuggestions: suggestions,
	}
}

// ServerGroup represents a server record with its related records
type ServerGroup struct {
	Server         models.DNSRecord   `json:"server"`
	RelatedRecords []models.DNSRecord `json:"related_records"`
}

// UnassignedGroup represents unassigned records grouped by provider
type UnassignedGroup struct {
	ProviderID   uint               `json:"provider_id"`
	ProviderName string             `json:"provider_name"`
	Records      []models.DNSRecord `json:"records"`
}

// GroupedRecords represents the top-level grouping structure (server-first)
type GroupedRecords struct {
	Servers              []ServerGroup                 `json:"servers"`
	UnassignedRecords    []UnassignedGroup             `json:"unassigned_records"`
	ProviderCapabilities map[uint]ProviderCapabilities `json:"provider_capabilities"`
}

// GroupRecords groups DNS records by server first, then unassigned records by provider
// This supports cross-provider analysis where CNAMEs can point across providers
func GroupRecords(records []models.DNSRecord, providers []models.Provider) GroupedRecords {
	providerMap := make(map[uint]models.Provider)
	for _, p := range providers {
		providerMap[p.ID] = p
	}

	capabilities := make(map[uint]ProviderCapabilities)
	for _, p := range providers {
		capabilities[p.ID] = GetProviderCapabilities(p)
	}

	// Separate ALL servers and non-servers across all providers
	var allServers []models.DNSRecord
	var allOtherRecords []models.DNSRecord
	recordUsed := make(map[uint]bool) // Track which records have been assigned

	for _, record := range records {
		if record.IsServer {
			allServers = append(allServers, record)
		} else {
			allOtherRecords = append(allOtherRecords, record)
		}
	}

	// Group servers by IP to merge duplicates
	serversByIP := make(map[string][]models.DNSRecord)
	for _, server := range allServers {
		serversByIP[server.TargetValue] = append(serversByIP[server.TargetValue], server)
	}

	// Build server groups (top level)
	var serverGroups []ServerGroup

	// For each IP, select the best primary server and merge others
	for _, serversWithSameIP := range serversByIP {
		// Select primary server: prefer region-formatted domain with server_name
		var primaryServer models.DNSRecord
		var otherServers []models.DNSRecord

		// Sort to find best primary: prefer region-formatted domains
		bestScore := -1
		for _, server := range serversWithSameIP {
			score := 0

			// Check if domain matches region-number format (highest priority)
			// and the region code is valid (exists in RegionMap)
			matches := serverPattern.FindStringSubmatch(server.FullDomain)
			if len(matches) > 1 {
				regionCode := matches[1]
				if _, exists := RegionMap[regionCode]; exists {
					score += 10
				}
			}

			// Additional points for server metadata
			if server.ServerName != "" {
				score += 2
			}
			if server.ServerRegion != "" {
				score += 1
			}

			if score > bestScore {
				if bestScore >= 0 {
					otherServers = append(otherServers, primaryServer)
				}
				primaryServer = server
				bestScore = score
			} else {
				otherServers = append(otherServers, server)
			}
		}

		serverGroup := ServerGroup{
			Server: primaryServer,
		}

		// Add other same-IP servers as related A records
		for _, otherServer := range otherServers {
			serverGroup.RelatedRecords = append(serverGroup.RelatedRecords, otherServer)
			recordUsed[otherServer.ID] = true
		}

		// Find records that point to this server (CNAME or same IP) from ALL providers
		for _, rec := range allOtherRecords {
			if recordUsed[rec.ID] {
				continue // Already assigned to another server
			}

			isRelated := false

			// Check if CNAME points to this primary server or any merged server
			if rec.RecordType == models.RecordTypeCNAME {
				target := rec.TargetValue
				// Remove trailing dot if present
				if len(target) > 0 && target[len(target)-1] == '.' {
					target = target[:len(target)-1]
				}

				// Check if points to primary server
				if target == primaryServer.FullDomain {
					isRelated = true
				} else {
					// Check if points to any of the merged servers
					for _, otherServer := range otherServers {
						if target == otherServer.FullDomain {
							isRelated = true
							break
						}
					}
				}
			}

			// Check if A record points to same IP (shouldn't happen as they're already merged, but keep for safety)
			if rec.RecordType == models.RecordTypeA && rec.TargetValue == primaryServer.TargetValue {
				isRelated = true
			}

			if isRelated {
				serverGroup.RelatedRecords = append(serverGroup.RelatedRecords, rec)
				recordUsed[rec.ID] = true
			}
		}

		serverGroups = append(serverGroups, serverGroup)
	}

	// Build unassigned records grouped by provider
	var unassignedGroups []UnassignedGroup
	providerUnassigned := make(map[uint][]models.DNSRecord)

	for _, rec := range allOtherRecords {
		if !recordUsed[rec.ID] {
			providerUnassigned[rec.ProviderID] = append(providerUnassigned[rec.ProviderID], rec)
		}
	}

	// Convert map to array
	for providerID, records := range providerUnassigned {
		provider := providerMap[providerID]
		unassignedGroups = append(unassignedGroups, UnassignedGroup{
			ProviderID:   providerID,
			ProviderName: provider.Name,
			Records:      records,
		})
	}

	// Sort server groups by domain name for consistent ordering
	sort.Slice(serverGroups, func(i, j int) bool {
		return serverGroups[i].Server.FullDomain < serverGroups[j].Server.FullDomain
	})

	// Sort related records within each server group
	for i := range serverGroups {
		sort.Slice(serverGroups[i].RelatedRecords, func(a, b int) bool {
			return serverGroups[i].RelatedRecords[a].FullDomain < serverGroups[i].RelatedRecords[b].FullDomain
		})
	}

	// Sort unassigned groups by provider name
	sort.Slice(unassignedGroups, func(i, j int) bool {
		return unassignedGroups[i].ProviderName < unassignedGroups[j].ProviderName
	})

	// Sort records within each unassigned group
	for i := range unassignedGroups {
		sort.Slice(unassignedGroups[i].Records, func(a, b int) bool {
			return unassignedGroups[i].Records[a].FullDomain < unassignedGroups[i].Records[b].FullDomain
		})
	}

	return GroupedRecords{
		Servers:              serverGroups,
		UnassignedRecords:    unassignedGroups,
		ProviderCapabilities: capabilities,
	}
}

// GetProviderCapabilities returns capability flags for a provider
func GetProviderCapabilities(provider models.Provider) ProviderCapabilities {
	supportsStatusToggle := false
	switch provider.Name {
	case models.ProviderTencentCloud:
		supportsStatusToggle = true
	}

	return ProviderCapabilities{
		SupportsRecordStatusToggle: supportsStatusToggle,
	}
}
