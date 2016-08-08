package models

import (
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
	"net/http"
	. "project/modules/helpers"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
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

// Checks playlist names for suffix
func playlistTitleRegexpMatch(candidate string) bool {
	r := regexp.MustCompile(WATCH_LATER_PLAYLIST_REGEXP)
	return r.Match([]byte(candidate))
}

// Compares playlist names to find out which is greater
func compareWatchLaterPlaylistTitles(a, b string) bool {
	return a > b
}

func getNextWatchLaterServicePlaylistDigit(title string) (digit int64, err error) {
	digit = -1
	if len(
		strings.TrimSpace(
			title,
		),
	) == 0 {
		return digit, fmt.Errorf(`Need a name of a playlist`)
	}

	extracted_digit := strings.TrimSpace(
		strings.Replace(
			title, WATCH_LATER_PLAYLIST_NAME, ``, 1,
		),
	)

	if len(extracted_digit) > 0 {
		if d, d_err := strconv.ParseInt(extracted_digit, 10, 64); d_err == nil {
			digit = d + 1
		} else {
			return -1, d_err
		}
	} else {
		digit = 2
	}

	return
}

// Creates new playlist with a title for a customer
func createPlaylist(
	youtubeService *youtube.Service,
	title string,
) (
	playlist_id string,
	err error,
) {

	// Create a `watch later` playlist
	watch_later_playlist_create_request := youtubeService.Playlists.Insert(
		"id,snippet,status",
		&youtube.Playlist{
			Id: WATCH_LATER_PLAYLIST_NAME,
			Snippet: &youtube.PlaylistSnippet{
				Title:       title,
				Description: "Periodically moves items from broken watch later. To stop it visit https://watchlater.chib.me",
				Tags:        []string{"watch later service", "created automatically"},
			},
			Status: &youtube.PlaylistStatus{
				PrivacyStatus: "private",
			},
		},
	)
	watch_later_playlist, err := watch_later_playlist_create_request.Do()
	return watch_later_playlist.Id, err
}

func copyVideoToPlaylist(
	youtubeService *youtube.Service,
	destination,
	videoId string,
) (response *youtube.PlaylistItem, err error) {
	copy_video_request := youtubeService.PlaylistItems.Insert(
		"id,snippet,status",
		&youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				PlaylistId: destination,
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
	response, err = copy_video_request.Do()
	return response, err
}

func getAllVideosFromPlaylist(
	youtubeService *youtube.Service,
	token *Token,
	playlist_id string,
) (videos Videos, err error) {
	videoCollectStart := time.Now().UTC()
	fmt.Println(
		"Started",
		token.Profile.Name,
		videoCollectStart,
	)

	nextPageToken := ""
	for {
		playlistCall := youtubeService.PlaylistItems.List("snippet").
			PlaylistId(playlist_id).
			MaxResults(50).
			PageToken(nextPageToken)

		playlistResponse, err := playlistCall.Do()
		if err != nil {
			fmt.Println(
				fmt.Sprintf("%s %v", token.Profile.Name, err),
			)
			Config.RavenClient.CaptureError(
				err,
				map[string]string{
					"TokenId": token.Id,
					"Email":   token.Profile.Email,
				},
			)
			return videos, err
		} else {
			for _, playlistItem := range playlistResponse.Items {
				video := &Video{
					Id: playlistItem.Snippet.ResourceId.VideoId,
				}
				videos = append(
					videos,
					video,
				)
			}
		}
		nextPageToken = playlistResponse.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	fmt.Println(
		fmt.Sprintf(
			"The customer %s has %d synced videos. Collection operation took %s", token.Profile.Name,
			len(videos),
			time.Since(videoCollectStart),
		),
	)

	return videos, err
}

func extractISO8601Duration(
	iso_8601_duration string,
) (
	minutes_duration,
	seconds_duration time.Duration,
	err error,
) {
	err = fmt.Errorf(`No duration can be extracted`)

	re := regexp.MustCompile(ISO_8601_DURATION_REGEX)
	if re.MatchString(iso_8601_duration) {
		matched_groups := re.FindAllStringSubmatch(iso_8601_duration, 1)
		if len(matched_groups) > 0 {
			groups := matched_groups[0]

			if d, d_err := strconv.ParseInt(groups[1], 10, 64); d_err == nil {
				minutes_duration = time.Minute * time.Duration(d)
				err = nil
			} else {
				err = d_err
			}
			if err == nil {
				if d, d_err := strconv.ParseInt(groups[2], 10, 64); d_err == nil {
					seconds_duration = time.Second * time.Duration(d)
					err = nil
				} else {
					err = d_err
				}
			}
		}
	}

	return
}

// Copy from `Watch Later` to a playlist named `Watch Later`. Then clear a `Watch Later`
func ProcessWatchLater(
	client *http.Client,
	token *Token,
) (
	err error,
) {
	if youtubeService, err := youtube.New(client); err != nil {
		return err
	} else {
		duration_playlists := newDurationPlaylists()
		existing_videos := make([]string, 0)

		var (
			moved_videos               Videos
			latest_watchlater_playlist *Playlist
			watchlater_playlists       Playlists
		)

		// if user has deleted a playlist recently created by this function - create ne wand save it's id
		nextPageToken := ""
		for {
			playlists_call := youtubeService.Playlists.List("id, snippet").
				MaxResults(50).
				PageToken(nextPageToken).
				Mine(true)

			playlists_response, err := playlists_call.Do()
			if err != nil {
				return err
			}
			// Set the token to retrieve the next page of results
			// or exit the loop if all results have been retrieved.
			nextPageToken = playlists_response.NextPageToken

			for _, playlist := range playlists_response.Items {
				for title, duration_playlist := range duration_playlists {
					if playlist.Snippet.Title == title {
						duration_playlist.playlistId = playlist.Id
					}
				}

				// if we found a playlist named "Watch Later Service"
				// or "Watch Later Service 3"
				if playlistTitleRegexpMatch(playlist.Snippet.Title) {
					videos, err := getAllVideosFromPlaylist(
						youtubeService,
						token,
						playlist.Id,
					)
					if err == nil {
						for _, video := range videos {
							existing_videos = append(
								existing_videos,
								video.Id,
							)
						}
					}

					watchlater_playlists = append(
						watchlater_playlists,
						&Playlist{
							Id:         playlist.Id,
							Title:      playlist.Snippet.Title,
							VideoItems: videos,
						},
					)

				}
			}
			if nextPageToken == "" {
				break
			}
		}

		if len(watchlater_playlists) > 0 {
			sort.Sort(watchlater_playlists)

			latest_watchlater_playlist = watchlater_playlists[len(watchlater_playlists)-1]
		}

		// IF we do not have a playlist - create new
		if latest_watchlater_playlist == nil {
			// Create a `watch later` playlist
			if playlist_id, err := createPlaylist(
				youtubeService,
				WATCH_LATER_PLAYLIST_NAME,
			); err == nil {
				latest_watchlater_playlist = &Playlist{
					Id:    playlist_id,
					Title: WATCH_LATER_PLAYLIST_NAME,
				}
				watchlater_playlists = append(
					watchlater_playlists,
					latest_watchlater_playlist,
				)
			} else {
				Config.RavenClient.CaptureError(
					err,
					map[string]string{
						"TokenId": token.Id,
						"Email":   token.Profile.Email,
					},
				)
				return err
			}
		}

		// We must get customer's original `Watch Later` Playlist
		channels_call := youtubeService.Channels.List("contentDetails").Mine(true)
		channels_response, err := channels_call.Do()
		if err != nil {
			Config.RavenClient.CaptureError(
				err,
				map[string]string{
					"TokenId": token.Id,
					"Email":   token.Profile.Email,
				},
			)
			return err
		}

		watchLaterId := ""
		for _, channel := range channels_response.Items {
			watchLaterId = channel.ContentDetails.RelatedPlaylists.WatchLater
			nextPageToken := ""
			for {
				playlistCall := youtubeService.PlaylistItems.List("id,snippet").
					PlaylistId(watchLaterId).
					MaxResults(50).
					PageToken(nextPageToken)
				playlistResponse, err := playlistCall.Do()

				if err != nil {
					Config.RavenClient.CaptureError(
						err,
						map[string]string{
							"TokenId": token.Id,
							"Email":   token.Profile.Email,
						},
					)
					return err
				}

				// if a `Watch Later` contains elements - collect all of them and reverse resulting list
				for _, playlistItem := range playlistResponse.Items {
					videoId := playlistItem.Snippet.ResourceId.VideoId

					// If this video is not in a `Watch Later Service` playlist
					if _, ok := StringInSlice(
						videoId,
						existing_videos,
					); !ok {
						var (
							video *Video
						)

						// Retrive video duration
						// TODO: if we'll impact some usage QUOTAS Limitations - remove this snippet and video title
						videos_list_call := youtubeService.Videos.List("id,contentDetails,snippet").
							Id(videoId)
						if videolistResponse, err := videos_list_call.Do(); err == nil {
							if len(videolistResponse.Items) == 1 {
								item := videolistResponse.Items[0]
								if minutes, seconds, err := extractISO8601Duration(
									item.ContentDetails.Duration,
								); err == nil {
									video = &Video{
										Id:       videoId,
										Title:    item.Snippet.Title,
										Duration: minutes + seconds,
									}
								}
							}
						}

						// Copy video to `Watch Later` Playlist
						if _, err := copyVideoToPlaylist(
							youtubeService,
							latest_watchlater_playlist.Id,
							videoId,
						); err != nil {
							if strings.Contains(
								err.Error(), "playlistContainsMaximumNumberOfVideos",
							) {
								// We must create new playlist
								if digit, err := getNextWatchLaterServicePlaylistDigit(
									latest_watchlater_playlist.Title,
								); err == nil {
									new_title := fmt.Sprintf("%s %d", WATCH_LATER_PLAYLIST_NAME, digit)

									if created_playlist_id, err := createPlaylist(
										youtubeService,
										new_title,
									); err == nil {
										latest_watchlater_playlist = &Playlist{
											Id:    created_playlist_id,
											Title: new_title,
										}
										watchlater_playlists = append(
											watchlater_playlists,
											latest_watchlater_playlist,
										)

										if _, err := copyVideoToPlaylist(
											youtubeService,
											latest_watchlater_playlist.Id,
											videoId,
										); err != nil {
											Config.RavenClient.CaptureError(
												err,
												map[string]string{
													"TokenId": token.Id,
													"Email":   token.Profile.Email,
												},
											)
											return err
										} else {
											if video != nil {
												moved_videos = append(
													moved_videos,
													video,
												)
											}
										}
									}
								} else {
									Config.RavenClient.CaptureError(
										err,
										map[string]string{
											"TokenId": token.Id,
											"Email":   token.Profile.Email,
										},
									)
									return err
								}
							}

							if !strings.Contains(err.Error(), "googleapi: Error 403: Forbidden, playlistItemsNotAccessible") &&
								!strings.Contains(err.Error(), "googleapi: Error 404: Video not found., videoNotFound") {
								Config.RavenClient.CaptureError(
									err,
									map[string]string{
										"TokenId": token.Id,
										"Email":   token.Profile.Email,
									},
								)
							}

							if !strings.Contains(err.Error(), "googleapi: Error 404: Video not found., videoNotFound") {
								return err
							}
						} else {
							if video != nil {
								moved_videos = append(
									moved_videos,
									video,
								)
							}
						}

						// // Delete a video from system `Watch Later`
						// playlistitems_service := youtube.NewPlaylistItemsService(youtubeService)
						// delete_call := playlistitems_service.Delete(playlistItem.Id)
						// err = delete_call.Do()
						// if err != nil {
						// 	fmt.Println(">>>Error while watchlater cleanup", err.Error())
						// 	if err.Error() != "googleapi: Error 404: Playlist item not found., playlistItemNotFound" {
						// 		Config.RavenClient.CaptureError(
						// 			err,
						// 			map[string]string{
						// 				"TokenId":      token.Id,
						// 				"Email":        token.Profile.Email,
						// 				"playlistItem": playlistItem.Id,
						// 			},
						// 		)
						// 	}
						// }
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

		sort.Sort(moved_videos)
		for _, video := range moved_videos {
			for title, duration_playlist := range duration_playlists {
				if duration_playlist.checkMatches(video.Duration) {
					if len(duration_playlist.playlistId) == 0 {
						if playlist_id, err := createPlaylist(
							youtubeService,
							title,
						); err == nil {
							duration_playlist.playlistId = playlist_id
						} else {
							Config.RavenClient.CaptureError(
								err,
								map[string]string{
									"TokenId": token.Id,
									"Email":   token.Profile.Email,
								},
							)
						}
					}

					if len(duration_playlist.playlistId) > 0 {
						if _, err := copyVideoToPlaylist(
							youtubeService,
							duration_playlist.playlistId,
							video.Id,
						); err != nil {
							Config.RavenClient.CaptureError(
								err,
								map[string]string{
									"TokenId": token.Id,
									"Email":   token.Profile.Email,
								},
							)
						}
					}
				}
			}
		}
	}

	return
}
