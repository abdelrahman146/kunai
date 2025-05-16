package es

var IndexMapping = map[string]interface{}{
	"settings": map[string]interface{}{
		"analysis": map[string]interface{}{
			// ─── TOKENIZERS ──────────────────────────────────────────────────
			"tokenizer": map[string]interface{}{
				"code_ngram_tokenizer": map[string]interface{}{
					"type":        "edge_ngram",
					"min_gram":    3,
					"max_gram":    15,
					"token_chars": []string{"letter", "digit"},
				},
			},
			// ─── FILTERS ─────────────────────────────────────────────────────
			"filter": map[string]interface{}{
				"camel_case_split": map[string]interface{}{
					"type":                  "word_delimiter",
					"generate_word_parts":   true,
					"generate_number_parts": true,
					"split_on_case_change":  true,
					"split_on_numerics":     true,
					"preserve_original":     true,
				},
			},
			// ─── ANALYZERS ───────────────────────────────────────────────────
			"analyzer": map[string]interface{}{
				"path_analyzer": map[string]interface{}{
					"tokenizer": "path_hierarchy",
				},
				"code_analyzer": map[string]interface{}{
					"tokenizer": "pattern",
					"filter":    []string{"camel_case_split", "lowercase", "asciifolding"},
					"pattern":   "[^A-Za-z0-9_]",
				},
				"code_ngram_analyzer": map[string]interface{}{
					"tokenizer": "code_ngram_tokenizer",
					"filter":    []string{"lowercase"},
				},
			},
			// ─── NORMALIZERS ─────────────────────────────────────────────────
			"normalizer": map[string]interface{}{
				"lowercase_normalizer": map[string]interface{}{
					"type":   "custom",
					"filter": []string{"lowercase"},
				},
			},
		},
	},

	"mappings": map[string]interface{}{
		"dynamic": false,
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":     "text",
				"analyzer": "english",
				"fields": map[string]interface{}{
					"raw": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
						"normalizer":   "lowercase_normalizer",
					},
				},
			},
			"project": map[string]interface{}{
				"type":         "keyword",
				"ignore_above": 256,
				"normalizer":   "lowercase_normalizer",
			},
			"relPath": map[string]interface{}{
				"type":     "text",
				"analyzer": "path_analyzer",
				"fields": map[string]interface{}{
					"raw": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 512,
						"normalizer":   "lowercase_normalizer",
					},
				},
			},
			"extension": map[string]interface{}{
				"type":       "keyword",
				"normalizer": "lowercase_normalizer",
			},
			"language": map[string]interface{}{
				"type":       "keyword",
				"normalizer": "lowercase_normalizer",
			},
			"size": map[string]interface{}{
				"type": "long",
			},
			"updatedAt": map[string]interface{}{
				"type":   "date",
				"format": "strict_date_optional_time||epoch_millis",
			},
			"content": map[string]interface{}{
				"type":            "text",
				"analyzer":        "code_analyzer",
				"search_analyzer": "standard",
				"term_vector":     "with_positions_offsets",
				"fields": map[string]interface{}{
					"ngram": map[string]interface{}{
						"type":     "text",
						"analyzer": "code_ngram_analyzer",
					},
					"suggest": map[string]interface{}{
						"type": "completion",
						"contexts": []map[string]interface{}{
							{
								"name": "project",
								"type": "category",
								"path": "project",
							},
						},
					},
				},
			},
		},
	},
}
