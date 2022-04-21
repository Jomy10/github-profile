package githubApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

// Fetch all (private and public) repos from a user
func FetchRepos(usr, token string) []Repository {
	repos, link := fetchRepos(usr, token, "https://api.github.com/user/repos")
	for link != "" {
		newRepos, newLink := fetchRepos(usr, token, link)
		link = newLink
		repos = append(repos, newRepos...)
	}
	return repos
}

// Fetch one page of repos
//
// Returns the repositories on the page and the link to the next page
func fetchRepos(usr, token, url string) ([]Repository, string) {
	req, reqErr := http.NewRequest("GET", url, nil)
	if reqErr != nil {
		panic(reqErr)
	}

	req.SetBasicAuth(usr, token)

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		panic(respErr)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Decode json response
	var data []Repository
	json.Unmarshal(body, &data)

	// next page link
	linkHeader := resp.Header.Get("link")
	regex := regexp.MustCompile("<(.*?)>; rel=\"next\", .*")
	linkMatches := regex.FindStringSubmatch(linkHeader)
	if len(linkMatches) > 1 {
		return data, linkMatches[1]
	} else {
		return data, ""
	}
}

// Fetch all public repos from a user
func FetchPublicRepos(usr string) []Repository {
	panic("unimplemented")
}

// All relevant information for repositories
type Repository struct {
	Id          int
	Name        string
	Full_Name   string
	Private     bool
	Description string
	Fork        bool
}

func FetchLanguagesFrom(repo Repository, usr, token string) map[string]uint {
	return FetchLanguages(repo.Full_Name, usr, token)
}

// Fetch languages from a repository
//
// `repo` is the full repository name
// `usr` and `token` are optional.
func FetchLanguages(repo string, usr, token string) map[string]uint {
	req, reqErr := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/languages", repo), nil)
	if reqErr != nil {
		panic(reqErr)
	}

	req.SetBasicAuth(usr, token)

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		panic(respErr)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Decode json response
	var data map[string]uint

	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}

	return data
}

// Fetch a map of all languages and their size from a user
//
//  - `usr`, `token`: authentication
//  - `inlcudeForks`: wether to include forks in the list
//  - `includePrivate`: wether to include private repositories in the list
//  - `withForks`: a list of repositories (full names) to include when includeForks is false
//  - `withPrivate`: a list of repositories (full name) to include when includePrivate is false
//  - `withoutForks`: a list of repositories (full name) to exclude when includForks is true
//  - `withoutPrivate`: a list of repositories (full name) to exclude when includePrivatt is true
func FetchUserLanguages(usr, token string, includeForks, includePrivate bool, withForks, withPrivate []string, withoutForks, withoutPrivate []string) map[string]uint {
	repos := FetchRepos(usr, token)
	langs := make(map[string]uint)

	for _, repo := range repos {
		if !includeForks && repo.Fork {
			include := false
			for _, includeForkName := range withForks {
				if repo.Full_Name == includeForkName {
					include = true
				}
			}
			if !include {
				continue
			}
		}
		if includeForks && repo.Fork {
			include := true
			for _, excludeForkName := range withoutForks {
				if repo.Full_Name == excludeForkName {
					include = false
				}
			}
			if !include {
				continue
			}
		}
		if !includePrivate && repo.Private {
			include := false
			for _, includePrivateName := range withPrivate {
				if repo.Full_Name == includePrivateName {
					include = true
				}
			}
			if !include {
				continue
			}
		}
		if includePrivate && repo.Private {
			include := true
			for _, excludePrivateName := range withoutPrivate {
				if repo.Full_Name == excludePrivateName {
					include = false
				}
			}
			if !include {
				continue
			}
		}

		for lang, size := range FetchLanguagesFrom(repo, usr, token) {
			langs[lang] += size
		}
	}

	return langs
}

func FetchPublicUserLanguages(usr string, includeForks bool, withForks, withoutForks []string) map[string]uint {
	panic("unimplemented")
}

type pageInfo struct {
	endCursor   string
	hasNextPage bool
}

// Fetch repos contributed (commited, opened pull request or creatd repository)
func FetchReposContributedTo(token string) []string {
	// contribution types: [COMMIT, ISSUE, PULL_REQUEST, REPOSITORY]
	var repos []string
	var response map[string]interface{} = fetchReposContributedTo(token, "")
	repos = readReposContributedTo(response)
	for response != nil {
		if m, ok := response["data"].(map[string]interface{}); ok {
			if m, ok = m["viewer"].(map[string]interface{}); ok {
				if m, ok = m["repositoriesContributedTo"].(map[string]interface{}); ok {
					if b, k := m["pageInfo"].(pageInfo); k {
						if b.hasNextPage {
							response = fetchReposContributedTo(token, b.endCursor)
							repos = append(repos, readReposContributedTo(response)...)
						} else {
							response = nil
						}
					} else {
						response = nil
					}
				} else {
					response = nil
				}
			} else {
				response = nil
			}
		} else {
			response = nil
		}
	}

	return repos
}

// fetch + page
func fetchReposContributedTo(token, page string) map[string]interface{} {
	var query []byte
	if page == "" {
		query = []byte(`{"query": "{ viewer { repositoriesContributedTo(first: 100 contributionTypes: [COMMIT, PULL_REQUEST, REPOSITORY]) { totalCount nodes { nameWithOwner } pageInfo { endCursor hasNextPage } } } }" }`)
	} else {
		query = []byte(`{"query": "{ viewer { repositoriesContributedTo(first: 100 after: "` + page + `" contributionTypes: [COMMIT, PULL_REQUEST, REPOSITORY]) { totalCount nodes { nameWithOwner } pageInfo { endCursor hasNextPage } } } }" }`)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(query))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	jsonData := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&jsonData)
	if err != nil {
		panic(err)
	}

	return jsonData
}

func readReposContributedTo(response map[string]interface{}) (ret []string) {
	// No proper error handling. Else blocks shouldn't happen unless GitHub changes their api
	if m, ok := response["data"].(map[string]interface{}); ok {
		if m, ok = m["viewer"].(map[string]interface{}); ok {
			if m, ok = m["repositoriesContributedTo"].(map[string]interface{}); ok {
				if a, okay := m["nodes"].([]interface{}); okay {
					for _, repo := range a {
						if r, k := repo.(map[string]interface{}); k {
							if str, kk := r["nameWithOwner"].(string); kk {
								ret = append(ret, str)
							} else {
								fmt.Println("error at nameWithOwner")
							}
						} else {
							fmt.Println(("error at repo conversion"))
						}
					}
				} else {
					fmt.Println("error at nodes")
				}
			} else {
				fmt.Println("error at repositoriesContributedTo")
			}
		} else {
			fmt.Println("error at viewer")
		}
	} else {
		fmt.Println("error at data")
	}
	return
}
