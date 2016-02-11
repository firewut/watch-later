package models

import (
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
	"net/http"
)

// Get youtube config. Uses a Config passed to a module
func GetYoutubeConfig() (oauth2Cfg *oauth2.Config) {
	return &oauth2.Config{
		ClientID:     Config.OAUTH2ClientID,
		ClientSecret: Config.OAUTH2ClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  Config.OAUTH2AuthUri,
			TokenURL: Config.OAUTH2TokenUri,
		},
		RedirectURL: Config.OAUTH2RedirectUri,
	}
}

// Copy from `Watch Later` to a playlist named `Watch Later`. Then clear a `Watch Later`
func ProcessWatchLater(client *http.Client, token *Token) (moved_videos map[string]string, err error) {
	moved_videos = make(map[string]string)

	if youtubeService, err := youtube.New(client); err != nil {
		return moved_videos, err
	} else {
		var WATCH_LATER_PLAYLIST_ID string

		// search a playlist by it's id
		if len(token.WatchLaterPlaylistId) > 0 {
			playlists_call := youtubeService.Playlists.List("id").Id(token.WatchLaterPlaylistId)
			playlists_response, err := playlists_call.Do()
			if err != nil {
				Log.Warning(fmt.Sprintf("search a playlist by it's id. %v", err))
			} else {
				// if we have a playlist recently created by this function - proceed with it's id
				if len(playlists_response.Items) == 1 {
					WATCH_LATER_PLAYLIST_ID = token.WatchLaterPlaylistId
				}
			}
		}

		// if user has deleted a playlist recently created by this function - create ne wand save it's id
		if len(WATCH_LATER_PLAYLIST_ID) == 0 {
			nextPageToken := ""
			for {
				// -- get one `Watch Later` playlist
				playlists_call := youtubeService.Playlists.List("id, snippet").
					MaxResults(50).
					PageToken(nextPageToken).Mine(true)

				playlists_response, err := playlists_call.Do()
				if err != nil {
					Log.Error(fmt.Sprintf("Error making API call to list playlists: %v", err))
					return moved_videos, err
				}
				// Set the token to retrieve the next page of results
				// or exit the loop if all results have been retrieved.
				nextPageToken = playlists_response.NextPageToken

				for _, playlist := range playlists_response.Items {
					// if we found a playlist named "Watch later"
					if playlist.Snippet.Title == WATCH_LATER_PLAYLIST_NAME {
						WATCH_LATER_PLAYLIST_ID = playlist.Id
						// Say - we no need a nextPageToken :D
						nextPageToken = ""
					}
				}
				if nextPageToken == "" {
					break
				}
			}
		}

		// -- If a playlist named `Watch Later` does not exist - create it
		if len(WATCH_LATER_PLAYLIST_ID) == 0 {
			// Create a `watch later` playlist
			watch_later_playlist_create_request := youtubeService.Playlists.Insert(
				"id,snippet,status",
				&youtube.Playlist{
					Id: WATCH_LATER_PLAYLIST_NAME,
					Snippet: &youtube.PlaylistSnippet{
						Title:       WATCH_LATER_PLAYLIST_NAME,
						Description: "Periodically moves items from broken watch later",
						Tags:        []string{"watch later", "created automatically"},
					},
					Status: &youtube.PlaylistStatus{
						PrivacyStatus: "private",
					},
				},
			)
			watch_later_playlist, err := watch_later_playlist_create_request.Do()
			if err != nil {
				Log.Error(fmt.Sprintf("Error while creating a `watch later channel` %v", err))
				return moved_videos, err
			}
			WATCH_LATER_PLAYLIST_ID = watch_later_playlist.Id
		}

		// We should have an id of a `watch later` playlist
		if len(WATCH_LATER_PLAYLIST_ID) > 0 {
			if err := token.Patch(map[string]interface{}{
				"watch_later_playlist_id": WATCH_LATER_PLAYLIST_ID,
			}); err != nil {
				Log.Error(err)
				return moved_videos, err
			}
		} else {
			Log.Error(
				fmt.Sprintf(
					"Token <%s>. Can't find a playlist with id <%s>. Also <%s> can't be created",
					token.OAuth2Token.AccessToken,
					token.WatchLaterPlaylistId,
					WATCH_LATER_PLAYLIST_NAME,
				),
			)
			return moved_videos, err
		}

		// 2
		// Get all user's channels
		channels_call := youtubeService.Channels.List("contentDetails").Mine(true)
		channels_response, err := channels_call.Do()
		if err != nil {
			Log.Error(fmt.Sprintf("Error making API call to list channels: %v", err))
			return moved_videos, err
		}

		watchLaterId := ""
		for _, channel := range channels_response.Items {
			watchLaterId = channel.ContentDetails.RelatedPlaylists.WatchLater
			nextPageToken := ""
			// videos_in_a_playlist := make([]string, 0)
			// watch_later_playlist_items := make([]string, 0)
			for {
				playlistCall := youtubeService.PlaylistItems.List("id,snippet").
					PlaylistId(watchLaterId).
					MaxResults(50).
					PageToken(nextPageToken)

				playlistResponse, err := playlistCall.Do()

				if err != nil {
					Log.Error(fmt.Sprintf("Error fetching `Watch Later` items: %v", err))
					return moved_videos, err
				}

				// if a `Watch Later` contains elements - collect all of them and reverse resulting list
				for _, playlistItem := range playlistResponse.Items {
					videoId := playlistItem.Snippet.ResourceId.VideoId
					// videos_in_a_playlist = append(videos_in_a_playlist, videoId)
					// watch_later_playlist_items = append(watch_later_playlist_items, playlistItem.Id)

					// Copy video to watch_later_playlists
					copy_video_request := youtubeService.PlaylistItems.Insert(
						"id,snippet,status",
						&youtube.PlaylistItem{
							Snippet: &youtube.PlaylistItemSnippet{
								PlaylistId: WATCH_LATER_PLAYLIST_ID,
								ResourceId: &youtube.ResourceId{
									Kind:    "youtube#video",
									VideoId: videoId,
								},
							},
							Status: &youtube.PlaylistItemStatus{
								PrivacyStatus: "private",
							},
						},
					)
					response, err := copy_video_request.Do()
					if err != nil {
						Log.Error(fmt.Sprintf("Error while creating a `watch later channel` %v", err))
						return moved_videos, err
					} else {
						moved_videos[videoId] = response.Id
					}

					// Delete a video from `Watch Later`
					playlistitems_service := youtube.NewPlaylistItemsService(youtubeService)
					delete_call := playlistitems_service.Delete(playlistItem.Id)
					err = delete_call.Do()
					if err != nil {
						Log.Error(fmt.Sprintf("Error cleaning `Watch Later` element. %s", err))
					}

				}

				// Set the token to retrieve the next page of results
				// or exit the loop if all results have been retrieved.
				nextPageToken = playlistResponse.NextPageToken
				if nextPageToken == "" {
					break
				}
			}
		}
	}

	return
}
