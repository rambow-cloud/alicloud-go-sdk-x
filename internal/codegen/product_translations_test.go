package codegen

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewedTranslationsKeepSourceAndRuntimeSeparate(t *testing.T) {
	p := readPolicyProduct(t, "ecs")
	if len(p.Translations) != 9 {
		t.Fatal("expected nine source-bound summaries")
	}
	translated, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(translated["docs/products/ecs.documentation.json"], []byte(productJSON)) {
		t.Fatal("documentation report lost its generator ownership marker")
	}
	p.Translations = nil
	original, err := renderProduct(p)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range original {
		if name == "service/ecs/operations.gen.go" || name == "docs/products/ecs.documentation.json" || name == "docs/products/ecs.md" || name == "docs/products/ecs.zh-CN.md" {
			continue
		}
		if !bytes.Equal(data, translated[name]) {
			t.Fatal("translation changed non-prose output", name)
		}
	}
	if !bytes.Contains(translated["service/ecs/operations.gen.go"], []byte("Reviewed English translation")) || !bytes.Contains(translated["service/ecs/operations.gen.go"], []byte("Queries instance topology.")) {
		t.Fatal("missing reviewed translation")
	}
	withoutComments := func(data []byte) string {
		var result []string
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "//") {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n")
	}
	if withoutComments(original["service/ecs/operations.gen.go"]) != withoutComments(translated["service/ecs/operations.gen.go"]) {
		t.Fatal("translation changed executable operation code")
	}
	var report productDocCoverage
	if json.Unmarshal(translated["docs/products/ecs.documentation.json"], &report) != nil {
		t.Fatal("report decoding")
	}
	if report.EnglishTranslations != 9 || report.EnglishOperations != 371 || report.GoOperationComments != 380 || report.GoFieldComments != report.TotalFields || report.Operations["CreateStorageSet"].Status != "reviewed-english-translation" {
		t.Fatal("conflated coverage", report.EnglishTranslations, report.EnglishOperations)
	}
	sts, err := renderProduct(readPolicyProduct(t, "sts"))
	if err != nil {
		t.Fatal(err)
	}
	guide := sts["docs/products/sts.md"]
	if bytes.Contains(guide, []byte("there is no federation discovery")) || !bytes.Contains(guide, []byte("config.LoadDefaultConfig discovers OIDC")) || !bytes.Contains(guide, []byte("Live federation renewal remains NOT RUN")) {
		t.Fatal("stale federation guidance")
	}
}

func TestTranslationDriftFailsBeforeWrites(t *testing.T) {
	root := fullProductFixture(t)
	p := readPolicyProduct(t, "ecs")
	file := productTranslationFile{SchemaVersion: 1, Product: p.Product, SourceManifestSHA256: p.Provenance.SourceManifestSHA256, Operations: p.Translations}
	entry := file.Operations["CreateStorageSet"]
	entry.TextSHA256 = "changed"
	file.Operations["CreateStorageSet"] = entry
	data, err := json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "translations/ecs.json", data)
	if err := GenerateProducts(context.Background(), root, false, nil); err == nil {
		t.Fatal("accepted changed source")
	}
	if _, err := os.Stat(filepath.Join(root, "service")); !os.IsNotExist(err) {
		t.Fatal("invalid translation wrote output", err)
	}
}

func TestTranslationValidationRejectsUnreviewedChanges(t *testing.T) {
	for _, mutate := range []func(*productIR){
		func(p *productIR) {
			entry := p.Translations["CreateStorageSet"]
			entry.English = "仅有中文"
			p.Translations["CreateStorageSet"] = entry
		},
		func(p *productIR) {
			entry := p.Translations["CreateStorageSet"]
			entry.Source.Line++
			p.Translations["CreateStorageSet"] = entry
		},
		func(p *productIR) {
			entry := p.Translations["CreateStorageSet"]
			entry.Chinese = "修改原文"
			p.Translations["CreateStorageSet"] = entry
		},
		func(p *productIR) {
			entry := p.Translations["CreateStorageSet"]
			entry.Attribute = "example"
			p.Translations["CreateStorageSet"] = entry
		},
		func(p *productIR) { p.Translations["UnknownOperation"] = p.Translations["CreateStorageSet"] },
	} {
		p := readPolicyProduct(t, "ecs")
		mutate(&p)
		if _, err := renderProduct(p); err == nil {
			t.Fatal("accepted invalid translation")
		}
	}
}
