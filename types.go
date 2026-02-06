package main

type Config struct {
	HostnameMappers []HostnameMapper `json:"hostname-mappers"`
}

func (this Config) GetTitle(hostname string) (string, bool) {
	title := hostname
	var ok bool

	for _, mapper := range this.HostnameMappers {
		if mapper.Hostname == hostname {
			title = mapper.Title
			ok = true
			break
		}
	}

	return title, ok
}

type HostnameMapper struct {
	Hostname string `json:"hostname"`
	Title    string `json:"title"`
}
