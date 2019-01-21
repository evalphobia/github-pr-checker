package prchecker

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFilePath = "config.yml"

	// enviroment parameters
	defaultEnvConfigFile  = "GITHUB_PR_CONFIG_FILE"
	defaultEnvAPIToken    = "GITHUB_PR_API_TOKEN"
	defaultEnvAPITokenKMS = "GITHUB_PR_API_TOKEN_KMS"
)

var (
	envFileName    string
	envAPIToken    string
	envAPITokenKMS string
)

func init() {
	envFileName = os.Getenv(defaultEnvConfigFile)
	envAPIToken = os.Getenv(defaultEnvAPIToken)
	envAPITokenKMS = os.Getenv(defaultEnvAPITokenKMS)
}

// Config contains settings.
type Config struct {
	RepositoryList map[string]*Repository `yaml:"repository"`
	APIToken       string                 `yaml:"api_token"`
	APITokenKMS    string                 `yaml:"api_token_kms"`
	BotID          int64                  `yaml:"bot_id"`
	WebHookSecret  string                 `yaml:"webhook_secret"`
	// If `AddComment` is true, comment is created each push event.
	// If `AddComment` is false, comment is edited each push event.
	AddComment bool `yaml:"add_comment"`

	Timeout time.Duration `yaml:"timeout"`
}

// NewConfig creates Config from yaml file and environment paramter.
func NewConfig() (Config, error) {
	var conf Config
	var err error
	switch {
	case isEnvironmentConfigExists():
		conf, err = loadConfigFromFile(envFileName)
	case isDefaultConfigExists():
		conf, err = loadConfigFromFile(defaultConfigFilePath)
	}
	if err != nil {
		return conf, err
	}

	err = conf.Init()
	return conf, err
}

// Init initializes config.
func (c *Config) Init() error {
	// get api token through KMS
	if !c.HasAPIToken() {
		var err error
		switch {
		case c.APITokenKMS != "":
			c.APIToken, err = DecryptKMS(c.APITokenKMS)
		case envAPITokenKMS != "":
			c.APIToken, err = DecryptKMS(c.APITokenKMS)
		}
		if err != nil {
			return err
		}
	}

	// initialize repository list and file's regexp
	if c.RepositoryList == nil {
		c.RepositoryList = make(map[string]*Repository)
	}
	for _, r := range c.RepositoryList {
		r.Init()
	}
	return nil
}

// HasAPIToken checks if api token is exists or not.
func (c Config) HasAPIToken() bool {
	return c.GetAPIToken() != ""
}

// GetAPIToken gets api token for GitHub.
func (c Config) GetAPIToken() string {
	if c.APIToken != "" {
		return c.APIToken
	}
	return envAPIToken
}

// GetRepository gets Repository setting to manage pull request.
func (c Config) GetRepository(name string) *Repository {
	if c.RepositoryList == nil {
		return nil
	}

	r, ok := c.RepositoryList[name]
	if !ok {
		return nil
	}
	return r
}

func loadConfigFromFile(path string) (Config, error) {
	conf := Config{}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(buf, &conf)
	return conf, err
}

func isEnvironmentConfigExists() bool {
	return envFileName != "" && isFileExists(envFileName)
}

func isDefaultConfigExists() bool {
	return isFileExists(defaultConfigFilePath)
}

func isFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Repository has settings to check changed files.
type Repository struct {
	Files []*File
}

// Init initializes all of Files.
func (r *Repository) Init() {
	for _, f := range r.Files {
		f.Init()
	}
}

// File has rules to check changed file and manage comment/assignees/reviwers of pull request.
type File struct {
	once sync.Once `yaml:"-"`

	Name      string           `yaml:"name"`
	Regexp    []string         `yaml:"regexp"`
	regexp    []*regexp.Regexp `yaml:"-"`
	Comment   string           `yaml:"comment"`
	Assignees []string         `yaml:"assignees"`
	Reviewers []string         `yaml:"reviewers"`
	ShowFiles bool             `yaml:"show_files"`
}

// Init initializes regexp rules.
func (f *File) Init() {
	f.once.Do(func() {
		fmt.Printf("[Init] %s\n", f.Name)
		list := make([]*regexp.Regexp, len(f.Regexp))
		for i, r := range f.Regexp {
			list[i] = regexp.MustCompile(r)
		}
		f.regexp = list
	})
}

// Match checks file path matches the regexp rule or not..
func (f *File) Match(path string) bool {
	f.Init()
	for _, r := range f.regexp {
		isMatch := r.MatchString(path)
		if isMatch {
			return true
		}
	}
	return false
}

// GetComment gets comment.
func (f *File) GetComment(files []string) string {
	if f.ShowFiles {
		return fmt.Sprintf("[%s]\n%s\n- %s", f.Name, f.Comment, strings.Join(files, "\n- "))
	}
	return fmt.Sprintf("[%s]\n%s", f.Name, f.Comment)
}
