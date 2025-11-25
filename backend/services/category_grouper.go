package services

import (
	"strings"
)

type CategoryGrouper struct {
	rules       map[string]string   // specific term -> group
}

// categoryDefinitions - Simple map of category -> keywords
// To add a new category, just add an entry here with the category name and its keywords
var categoryDefinitions = map[string][]string{
	"doctor": {
		"cardiologist", "neurologist", "ent", "orthopedic", "pediatrician",
		"dermatologist", "psychiatrist", "surgeon", "physician", "doctor",
		"dentist", "orthodontist", "oncologist", "radiologist",
		"dr", "doc", "md", "medical doctor",
	},
	"software engineer": {
		"developer", "programmer", "software engineer", "coder",
		"frontend developer", "backend developer", "full stack developer",
		"web developer", "mobile developer", "devops engineer",
		"software engineering", "software development", "web development",
		"mobile development", "application development", "coding", "software",
		"dev", "swe", "se",
	},
	"lawyer": {
		"lawyer", "attorney", "advocate", "solicitor", "barrister",
		"legal counsel", "legal advisor",
		"atty", "esq", "legal",
	},
	"teacher": {
		"teacher", "professor", "instructor", "educator", "tutor",
		"lecturer", "trainer",
		"prof",
	},
	"manager": {
		"manager", "director", "executive", "ceo", "cto", "cfo",
		"president", "vp", "vice president", "team lead",
		"mgr", "supervisor", "lead",
	},
	"designer": {
		"designer", "graphic designer", "ui designer", "ux designer",
		"product designer", "artist", "illustrator",
		"branding", "brand", "visual design", "creative design",
		"ux", "ui", "graphic",
	},
	"sales professional": {
		"sales", "salesperson", "sales rep", "sales representative",
		"account executive", "business development", "marketing",
		"marketing manager", "brand manager",
		"copywriting", "positioning", "strategy", "insight",
		"sales rep", "ae", "bdm",
	},
	"accountant": {
		"accountant", "auditor", "financial analyst", "bookkeeper",
		"tax consultant", "chartered accountant", "cpa",
		"ca", "finance",
	},
	"engineer": {
		"engineer", "mechanical engineer", "civil engineer", "electrical engineer",
		"chemical engineer", "aerospace engineer", "industrial engineer",
		"environmental engineer", "biomedical engineer",
		"eng", "engr",
	},
	"healthcare professional": {
		"nurse", "pharmacist", "therapist", "physiotherapist",
		"paramedic", "medical assistant", "lab technician",
		"radiographer", "dietitian", "nutritionist",
		"rn", "lpn", "medical staff",
	},
	"construction worker": {
		"construction worker", "contractor", "builder", "carpenter",
		"electrician", "plumber", "mason", "welder",
		"tradesman",
	},
	"hospitality professional": {
		"chef", "cook", "waiter", "waitress", "bartender",
		"hotel manager", "receptionist", "concierge",
		"server", "hospitality",
	},
	"retail professional": {
		"cashier", "store manager", "retail assistant", "sales associate",
		"merchandiser", "stock clerk",
	},
	"transportation worker": {
		"driver", "truck driver", "delivery driver", "pilot",
		"captain", "logistics coordinator", "dispatcher",
	},
	"manufacturing worker": {
		"factory worker", "production supervisor", "assembly line worker",
		"quality inspector", "machine operator", "foreman",
	},
	"public servant": {
		"police officer", "firefighter", "government official",
		"civil servant", "social worker", "public administrator",
	},
	"media professional": {
		"journalist", "reporter", "editor", "writer", "author",
		"photographer", "videographer", "content creator",
	},
	"researcher": {
		"scientist", "researcher", "analyst", "data scientist",
		"biologist", "chemist", "physicist", "research assistant",
		"research", "analysis", "data analysis", "scientific research",
	},
	"hr professional": {
		"hr", "human resources", "recruiter", "talent acquisition",
		"hr manager", "hr specialist", "hiring manager",
		"recruitment", "talent",
	},
	"security professional": {
		"security", "cybersecurity", "information security", "network security",
		"security analyst", "security engineer", "homeland security",
		"airport security", "screening", "explosive detection",
		"cargo screening", "checkpoint screening", "baggage screening",
	},
	"technology specialist": {
		"technology", "advanced technology", "innovation", "tech",
		"it specialist", "systems analyst", "network administrator",
		"database administrator", "cloud computing", "ai", "machine learning",
		"computed tomography", "ct scan", "imaging technology",
	},
	"internet professional": {
		"internet", "internet marketing", "digital marketing", "digital",
		"seo", "sem", "search engine optimization", "search engine marketing",
		"online marketing", "web marketing", "ecommerce", "e-commerce",
		"social media marketing", "content marketing", "email marketing",
		"technicalseo", "technical seo", "digitaltransformation", "digital transformation",
		"marketingautomation", "marketing automation", "crm",
	},
	"drawings": {
		"drawings", "sketch", "blueprint", "draft", "illustration",
		"diagram",
	},
	"design": {
		"design",
	},
}

func NewCategoryGrouper() *CategoryGrouper {
	grouper := &CategoryGrouper{
		rules:      make(map[string]string),
	}
	grouper.initializeRules()
	return grouper
}

// initializeRules builds the rules map from categoryDefinitions
func (g *CategoryGrouper) initializeRules() {
	for category, keywords := range categoryDefinitions {
		for _, keyword := range keywords {
			g.rules[strings.ToLower(keyword)] = category
		}
	}
}

// levenshteinDistance calculates the minimum edits needed between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// GetGroup returns the unified group for a given category with intelligent matching
func (g *CategoryGrouper) GetGroup(category string) string {
	cleaned := strings.ToLower(strings.TrimSpace(category))
	
	// Empty check
	if cleaned == "" {
		return ""
	}

	// 1. Direct match
	if group, ok := g.rules[cleaned]; ok {
		return group
	}

	// 2. Partial match - check if any keyword is a complete word in the category
	for key, group := range g.rules {
		if strings.Contains(" "+cleaned+" ", " "+key+" ") {
			return group
		}
	}

	// 3. Limited fuzzy match - only for very close matches (1 character difference, typos only)
	bestMatch := ""
	bestDistance := 999
	maxDistance := 1 // Only allow 1 character difference

	for key, group := range g.rules {
		// Only fuzzy match if lengths are very similar and string is reasonably long
		if abs(len(cleaned)-len(key)) <= 1 && len(cleaned) >= 5 {
			distance := levenshteinDistance(cleaned, key)
			if distance < bestDistance && distance <= maxDistance {
				bestDistance = distance
				bestMatch = group
			}
		}
	}

	if bestMatch != "" {
		return bestMatch
	}

	// No match found
	return ""
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// AddRule allows dynamic addition of grouping rules
func (g *CategoryGrouper) AddRule(term string, group string) {
	g.rules[strings.ToLower(term)] = group
}

// GetAllGroups returns all defined groups with their keywords
func (g *CategoryGrouper) GetAllGroups() map[string][]string {
	// Return a copy of categoryDefinitions
	result := make(map[string][]string)
	for category, keywords := range categoryDefinitions {
		result[category] = append([]string{}, keywords...)
	}
	return result
}
