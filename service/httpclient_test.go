package service

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestSearch(t *testing.T) {
	_, err := Search("kotlin")
	assert.Nil(t, err)
}

func TestArtifact(t *testing.T) {
	artifactResults, _ := Artifact("org.jetbrains.kotlin", "kotlin-stdlib")
	for _, v := range artifactResults {
		log.Printf("%v", v)
	}
}

func TestVersion(t *testing.T) {
	versionResult, _ := Version("org.jetbrains.kotlin", "kotlin-stdlib", "1.3.11")
	log.Printf("%v", versionResult)
}
