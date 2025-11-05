// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

// Pipeline provides the pipeline environment.
type Pipeline struct {
	Build  Build  `json:"build"`
	Repo   Repo   `json:"repo"`
	Stage  Stage  `json:"stage"`
	System System `json:"system"`
}

// Build provides the current build environment.
type Build struct {
	Branch      string `json:"branch"`
	Action      string `json:"action"`
	Number      int    `json:"number"`
	Parent      int    `json:"parent"`
	Event       string `json:"event"`
	Status      string `json:"status"`
	Deploy      string `json:"deploy"`
	Created     int64  `json:"created"`
	Started     int64  `json:"started"`
	Finished    int64  `json:"finished"`
	Link        string `json:"link"`
	Target      string `json:"target"`
	Commit      Commit `json:"commit"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	Source      string `json:"source"`
	Author      Author `json:"author"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	AuthorLogin string `json:"author_login"`
	Sender      string `json:"sender"`
	Params      map[string]string
}

// Commit provides the current commit environment.
type Commit struct {
	Remote  string `json:"remote"`
	Sha     string `json:"sha"`
	Ref     string `json:"ref"`
	Link    string `json:"link"`
	Branch  string `json:"branch"`
	Message string `json:"message"`
	Author  Author `json:"author"`
}

// Author provides the current author environment.
type Author struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Login  string `json:"login"`
	Avatar string `json:"avatar"`
}

// Repo provides the current repository environment.
type Repo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Slug      string `json:"slug"`
	SCM       string `json:"scm"`
	HTTPURL   string `json:"http_url"`
	SSHURL    string `json:"ssh_url"`
	Link      string `json:"link"`
	Branch    string `json:"branch"`
	Private   bool   `json:"private"`
	Trusted   bool   `json:"trusted"`
}

// Stage provides the current stage environment.
type Stage struct {
	Kind      string            `json:"kind"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Number    int               `json:"number"`
	Machine   string            `json:"machine"`
	OS        string            `json:"os"`
	Arch      string            `json:"arch"`
	Variant   string            `json:"variant"`
	Version   string            `json:"version"`
	Vendor    string            `json:"vendor"`
	Started   int64             `json:"started"`
	Finished  int64             `json:"finished"`
	Created   int64             `json:"created"`
	Updated   int64             `json:"updated"`
	Status    string            `json:"status"`
	ExitCode  int               `json:"exit_code"`
	DependsOn []string          `json:"depends_on"`
	Labels    map[string]string `json:"labels"`
}

// System provides the current system environment.
type System struct {
	Proto   string `json:"proto"`
	Host    string `json:"host"`
	Link    string `json:"link"`
	Version string `json:"version"`
}
