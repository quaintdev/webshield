package service

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"github.com/quaintdev/webshield/src/internal/repository"
)

type CertConf struct {
	CertPath string
	KeyPath  string
}

type Category struct {
	Name     string `json:"name"`
	FilePath string `json:"file"`
}

type Config struct {
	CertConfig        CertConf
	Categories        []Category
	DNSServers        []string
	WebsiteExceptions []Category
}

type ApplicationConfigService struct {
	config *Config
}

func NewApplicationConfigService(path string) *ApplicationConfigService {
	confFile, err := os.Open(path)
	if err != nil {
		slog.Error("error opening config file", "error", err)
		return nil
	}
	defer confFile.Close()

	var config *Config
	err = json.NewDecoder(confFile).Decode(&config)
	if err != nil {
		slog.Error("error parsing config file", "error", err)
		return nil
	}
	return &ApplicationConfigService{
		config: config,
	}
}

func (c *ApplicationConfigService) LoadDomainData(domainRepo repository.DomainDataRepository) {
	for _, category := range c.config.Categories {
		slog.Debug("loading blocklist", "name", category.Name)
		file, err := os.Open(category.FilePath)
		if err != nil {
			slog.Error("failed to open category file", "error", err)
			return
		}
		defer file.Close()

		if os.Getenv("TEST") == "true" {
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(strings.TrimSpace(line), "#") {
				domainRepo.AddDomain(line, category.Name)
			}
		}
		if err := scanner.Err(); err != nil {
			slog.Error("failed to read category file", "error", err)
		}
	}
}

func (c *ApplicationConfigService) GetCertConf() *CertConf {
	return &c.config.CertConfig
}

func (c *ApplicationConfigService) GetDNSServers() []string {
	return c.config.DNSServers
}

func (c *ApplicationConfigService) GetCategories() []Category {
	return c.config.Categories
}
