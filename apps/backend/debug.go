package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test needsModelImport logic
	types := []struct{
		Name string
		Fields []struct{
			Type string
		}
	}{
		{
			Name: "GetLinkedPubkeysResponse", 
			Fields: []struct{Type string}{
				{Type: "LinkedPubkeyInfo[]"},
			},
		},
	}
	
	for _, t := range types {
		for _, field := range t.Fields {
			cleanType := strings.TrimPrefix(strings.TrimPrefix(field.Type, "[]"), "*")
			fmt.Printf("Field type: %s, Clean type: %s, Is model: %t
", field.Type, cleanType, isModelType(cleanType))
		}
	}
}

func isModelType(typeName string) bool {
	modelTypes := []string{
		"User", "Track", "APIUser", "NostrAuth", "LinkedPubkeyInfo", 
		"CompressionOption", "CompressionVersion", "NostrTrack", "VersionUpdate",
		"LegacyUser", "LegacyTrack", "LegacyArtist", "LegacyAlbum",
	}
	
	for _, modelType := range modelTypes {
		if typeName == modelType {
			return true
		}
	}
	return false
}
