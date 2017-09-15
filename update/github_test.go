//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-03
//

package update

import (
	"fmt"
	"testing"

	aw "github.com/deanishe/awgo"
)

var (
	ghReleasesEmptyJSON = `[]`
	// 4 valid releases, including one prerelease
	// v1.0, v2.0, v6.0 and v7.1.0-beta
	ghReleasesJSON = `
[
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/617375/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v7.1.0-beta",
    "id": 617375,
    "tag_name": "v7.1.0-beta",
    "target_commitish": "master",
    "name": "Invalid release (pre-release status)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": true,
    "created_at": "2014-10-10T10:58:14Z",
    "published_at": "2014-10-10T10:59:34Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/265007",
        "id": 265007,
        "name": "Dummy-7.1-beta.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 35726,
        "download_count": 4,
        "created_at": "2014-10-10T10:59:10Z",
        "updated_at": "2014-10-10T10:59:12Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v7.1.0-beta/Dummy-7.1-beta.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v7.1.0-beta",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v7.1.0-beta",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556526/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v7.0",
    "id": 556526,
    "tag_name": "v7.0",
    "target_commitish": "master",
    "name": "Invalid release (contains no files)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T19:25:55Z",
    "published_at": "2014-09-14T19:27:25Z",
    "assets": [

    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v7.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v7.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556525/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v6.0",
    "id": 556525,
    "tag_name": "v6.0",
    "target_commitish": "master",
    "name": "Latest valid release",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T19:24:55Z",
    "published_at": "2014-09-14T19:27:09Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/4823231",
        "id": 4823231,
        "name": "Dummy-6.0.alfred3workflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2017-09-14T12:22:03Z",
        "updated_at": "2017-09-14T12:22:08Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfred3workflow"
      },
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247310",
        "id": 247310,
        "name": "Dummy-6.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 584,
        "created_at": "2014-09-23T18:59:00Z",
        "updated_at": "2014-09-23T18:59:01Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.alfredworkflow"
      },
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247311",
        "id": 247311,
        "name": "Dummy-6.0.zip",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/zip",
        "state": "uploaded",
        "size": 36063,
        "download_count": 1,
        "created_at": "2014-09-23T18:59:00Z",
        "updated_at": "2014-09-23T18:59:01Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v6.0/Dummy-6.0.zip"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v6.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v6.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556524/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v5.0",
    "id": 556524,
    "tag_name": "v5.0",
    "target_commitish": "master",
    "name": "Invalid release (contains no files)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T19:22:44Z",
    "published_at": "2014-09-14T19:26:30Z",
    "assets": [

    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v5.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v5.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556356/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v4.0",
    "id": 556356,
    "tag_name": "v4.0",
    "target_commitish": "master",
    "name": "Invalid release (contains 2 .alfredworkflow files)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:34:44Z",
    "published_at": "2014-09-14T16:36:34Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247308",
        "id": 247308,
        "name": "Dummy-4.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 693,
        "created_at": "2014-09-23T18:58:25Z",
        "updated_at": "2014-09-23T18:58:27Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v4.0/Dummy-4.0.alfredworkflow"
      },
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247309",
        "id": 247309,
        "name": "Dummy-4.1.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2014-09-23T18:58:26Z",
        "updated_at": "2014-09-23T18:58:27Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v4.0/Dummy-4.1.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v4.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v4.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556354/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v3.0",
    "id": 556354,
    "tag_name": "v3.0",
    "target_commitish": "master",
    "name": "Invalid release (no .alfredworkflow file)",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:34:16Z",
    "published_at": "2014-09-14T16:36:16Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247305",
        "id": 247305,
        "name": "Dummy-3.0.zip",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/zip",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2014-09-23T18:57:53Z",
        "updated_at": "2014-09-23T18:57:54Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v3.0/Dummy-3.0.zip"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v3.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v3.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556352/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v2.0",
    "id": 556352,
    "tag_name": "v2.0",
    "target_commitish": "master",
    "name": "",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:33:36Z",
    "published_at": "2014-09-14T16:35:47Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247300",
        "id": 247300,
        "name": "Dummy-2.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2014-09-23T18:57:19Z",
        "updated_at": "2014-09-23T18:57:21Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v2.0/Dummy-2.0.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v2.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v2.0",
    "body": ""
  },
  {
    "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350",
    "assets_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350/assets",
    "upload_url": "https://uploads.github.com/repos/deanishe/alfred-workflow-dummy/releases/556350/assets{?name,label}",
    "html_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/tag/v1.0",
    "id": 556350,
    "tag_name": "v1.0",
    "target_commitish": "master",
    "name": "",
    "draft": false,
    "author": {
      "login": "deanishe",
      "id": 747913,
      "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/deanishe",
      "html_url": "https://github.com/deanishe",
      "followers_url": "https://api.github.com/users/deanishe/followers",
      "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
      "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
      "organizations_url": "https://api.github.com/users/deanishe/orgs",
      "repos_url": "https://api.github.com/users/deanishe/repos",
      "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
      "received_events_url": "https://api.github.com/users/deanishe/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2014-09-14T16:33:06Z",
    "published_at": "2014-09-14T16:35:25Z",
    "assets": [
      {
        "url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/releases/assets/247299",
        "id": 247299,
        "name": "Dummy-1.0.alfredworkflow",
        "label": null,
        "uploader": {
          "login": "deanishe",
          "id": 747913,
          "avatar_url": "https://avatars1.githubusercontent.com/u/747913?v=4",
          "gravatar_id": "",
          "url": "https://api.github.com/users/deanishe",
          "html_url": "https://github.com/deanishe",
          "followers_url": "https://api.github.com/users/deanishe/followers",
          "following_url": "https://api.github.com/users/deanishe/following{/other_user}",
          "gists_url": "https://api.github.com/users/deanishe/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/deanishe/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/deanishe/subscriptions",
          "organizations_url": "https://api.github.com/users/deanishe/orgs",
          "repos_url": "https://api.github.com/users/deanishe/repos",
          "events_url": "https://api.github.com/users/deanishe/events{/privacy}",
          "received_events_url": "https://api.github.com/users/deanishe/received_events",
          "type": "User",
          "site_admin": false
        },
        "content_type": "application/octet-stream",
        "state": "uploaded",
        "size": 36063,
        "download_count": 0,
        "created_at": "2014-09-23T18:56:22Z",
        "updated_at": "2014-09-23T18:56:24Z",
        "browser_download_url": "https://github.com/deanishe/alfred-workflow-dummy/releases/download/v1.0/Dummy-1.0.alfredworkflow"
      }
    ],
    "tarball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/tarball/v1.0",
    "zipball_url": "https://api.github.com/repos/deanishe/alfred-workflow-dummy/zipball/v1.0",
    "body": ""
  }
]
`
)

func TestParseGH(t *testing.T) {
	rels, err := parseGitHubReleases([]byte(ghReleasesEmptyJSON))
	if err != nil {
		t.Fatal("Error parsing empty JSON.")
	}
	if len(rels) != 0 {
		t.Fatal("Found releases in empty JSON.")
	}
	rels, err = parseGitHubReleases([]byte(ghReleasesJSON))
	if err != nil {
		t.Fatal("Couldn't parse GitHub JSON.")
	}
	if len(rels) != 4 {
		t.Fatalf("Found %d GitHub releases, not 4.", len(rels))
	}
}

// makeGHReleaser creates a new GitHub Releaser and populates its release cache.
func makeGHReleaser() *GitHubReleaser {
	gh := &GitHubReleaser{Repo: "deanishe/nonexistent"}
	// Avoid network
	rels, _ := parseGitHubReleases([]byte(ghReleasesJSON))
	gh.releases = rels
	return gh
}

func TestGHUpdater(t *testing.T) {
	v := &versioned{version: "0.2.2"}
	defer v.Clean()
	gh := makeGHReleaser()

	// There are 4 valid releases (one prerelease)
	rels, err := gh.Releases()
	if err != nil {
		t.Fatalf("Error retrieving GH releases: %s", err)
	}
	if len(rels) != 4 {
		t.Fatalf("Found %d valid releases, not 4.", len(rels))
	}

	// v6.0 is available
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
	u, err := New(v, gh)
	if err != nil {
		t.Fatalf("Error creating updater: %s", err)
	}
	u.CurrentVersion = mustVersion("2")

	// Update releases
	if err := u.CheckForUpdate(); err != nil {
		t.Fatalf("Couldn't retrieve releases: %s", err)
	}

	if !u.UpdateAvailable() {
		t.Fatal("No update found")
	}
	// v6.0 is the latest stable version
	u.CurrentVersion = mustVersion("6")
	if u.UpdateAvailable() {
		t.Fatal("Unexpectedly found update")
	}
	// Prerelease v7.1.0-beta is newer
	u.Prereleases = true
	if !u.UpdateAvailable() {
		t.Fatal("No update found")
	}
	if err := clearUpdateCache(); err != nil {
		t.Fatal(err)
	}
}

// TestUpdates ensures an unconfigured workflow doesn't think it can update
func TestUpdates(t *testing.T) {
	wf := aw.New()
	if err := wf.ClearCache(); err != nil {
		t.Fatal(fmt.Sprintf("couldn't clear cache: %v", err))
	}
	if wf.UpdateCheckDue() != false {
		t.Fatal("Unconfigured workflow wants to update")
	}
	if wf.UpdateAvailable() != false {
		t.Fatal("Unconfigured workflow wants to update")
	}
	if err := wf.CheckForUpdate(); err == nil {
		t.Fatal("Unconfigured workflow didn't error on update check")
	}
	if err := wf.InstallUpdate(); err == nil {
		t.Fatal("Unconfigured workflow didn't error on update install")
	}

	// Once more with an updater
	wf = aw.New(GitHub("deanishe/alfred-ssh"))
	if wf.UpdateCheckDue() != true {
		t.Fatal("Workflow doesn't want to update")
	}
	if err := wf.ClearCache(); err != nil {
		t.Fatal(err)
	}
}

// Configure Workflow to update from a GitHub repo.
func ExampleGitHub() {
	// Set source repo using GitHub Option
	wf := aw.New(GitHub("deanishe/alfred-ssh"))
	// Is a check for a newer version due?
	fmt.Println(wf.UpdateCheckDue())
	// Output:
	// true
}
