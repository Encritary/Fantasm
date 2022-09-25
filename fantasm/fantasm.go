package fantasm

import (
	"context"
	"encoding/json"
	vkapi "github.com/himidori/golang-vk-api"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

type Fantasm struct {
	log *logrus.Logger

	client *vkapi.VKGroupBot

	cancelCtx  context.Context
	cancelFunc context.CancelFunc

	sections       map[string]*Section
	defaultSection *Section

	photosCache map[string]*vkapi.PhotoAttachment

	userSectionsCache     *cache.Cache
	userPrevSectionsCache *cache.Cache
}

func NewFastasm(log *logrus.Logger, token string, sections map[string]*Section) (*Fantasm, error) {
	fsm := &Fantasm{
		log: log,

		sections:    sections,
		photosCache: make(map[string]*vkapi.PhotoAttachment),

		userSectionsCache:     cache.New(10*time.Minute, 15*time.Minute),
		userPrevSectionsCache: cache.New(10*time.Minute, 15*time.Minute),
	}

	fsm.cancelCtx, fsm.cancelFunc = context.WithCancel(context.Background())

	client, err := vkapi.NewVKGroupBot(token, &vkapi.TokenOptions{}, true)
	if err != nil {
		return nil, err
	}
	fsm.client = client

	fsm.defaultSection = fsm.sections["start"]
	return fsm, nil
}

func (fsm *Fantasm) AddSection(section *Section) {
	fsm.sections[section.ID] = section
}

func (fsm *Fantasm) Run() {
	fsm.client.AddBotsLongpollCallback("message_new", fsm.onNewMessage)
	fsm.client.ListenBotsLongPollServerWithCancel(fsm.cancelCtx)
}

func (fsm *Fantasm) onNewMessage(obj *vkapi.BotsLongPollObject) {
	section := fsm.defaultSection

	fsm.log.Info("Received message")

	userIDStr := strconv.FormatInt(obj.Message.SendByID, 10)

	var err error

	var secID string
	if rawSecID, ok := fsm.userSectionsCache.Get(userIDStr); ok {
		secID = rawSecID.(string)
	} else {
		secID, err = fsm.storageGet(obj.Message.SendByID, "fantasmCurrentSection")
		if err != nil {
			panic(err)
		}
	}

	if secID != "" {
		if sec, ok := fsm.sections[secID]; ok {
			section = sec
		}
	}

	fsm.log.Debug("Parsing message")

	var action *Action
	if len(obj.Message.Text) > 1 && obj.Message.Text[0] == '/' {
		num, err := strconv.Atoi(obj.Message.Text[1:])

		if err == nil && num > 0 && num <= len(section.Actions) {
			action = section.Actions[num-1]
		}
	} else if obj.Message.Payload != "" {
		var payload map[string]interface{}
		err := json.Unmarshal([]byte(obj.Message.Payload), &payload)
		if err != nil {
			panic(err)
		}

		if rawNum, ok := payload["action"]; ok {
			if num, ok := rawNum.(float64); ok {
				action = section.Actions[int(num)]
			}
		}
	}

	var toSelect *Section
	if action != nil {
		if action.Section == "__back" {
			var prevSecID string

			if rawSecID, ok := fsm.userPrevSectionsCache.Get(userIDStr); ok {
				prevSecID = rawSecID.(string)
			} else {
				prevSecID, err = fsm.storageGet(obj.Message.SendByID, "fantasmPrevSection")
				if err != nil {
					panic(err)
				}
			}

			if sec, ok := fsm.sections[prevSecID]; ok {
				toSelect = sec
			} else {
				toSelect = fsm.defaultSection
			}
		} else if sec, ok := fsm.sections[action.Section]; ok {
			toSelect = sec
		}
	}

	if toSelect != nil {
		if section.ID == fsm.defaultSection.ID {
			err = fsm.storageSet(obj.Message.SendByID, "fantasmPrevSection", "")
			if err != nil {
				panic(err)
			}
			fsm.userPrevSectionsCache.Delete(userIDStr)
		} else {
			err = fsm.storageSet(obj.Message.SendByID, "fantasmPrevSection", section.ID)
			if err != nil {
				panic(err)
			}
			fsm.userPrevSectionsCache.Set(userIDStr, section.ID, cache.DefaultExpiration)
		}

		err = fsm.storageSet(obj.Message.SendByID, "fantasmCurrentSection", toSelect.ID)
		if err != nil {
			panic(err)
		}
		fsm.userSectionsCache.Set(userIDStr, toSelect.ID, cache.DefaultExpiration)
		section = toSelect
	}

	params := paramsRandomID()
	data, err := json.Marshal(section.BuildKeyboard(false))
	if err != nil {
		panic(err)
	}
	params.Set("keyboard", string(data))

	if len(section.Images) > 0 {
		fsm.log.Debug("Uploading images")

		var toUpload []string
		var photos []*vkapi.PhotoAttachment

		for _, file := range section.Images {
			if photo, ok := fsm.photosCache[file]; ok {
				fsm.log.Debug("Got ", file, " from cache")
				photos = append(photos, photo)
			} else {
				fsm.log.Debug("Uploading ", file)
				toUpload = append(toUpload, file)
			}
		}

		if len(toUpload) > 0 {
			upl, err := fsm.client.UploadMessagesPhotos(int(obj.Message.PeerID), toUpload)

			if err != nil {
				panic(err)
			}

			for k, file := range toUpload {
				fsm.photosCache[file] = upl[k]
			}

			photos = append(photos, upl...)
		}

		params.Set("attachment", fsm.client.GetPhotosString(photos))
	}

	text := section.Text
	text += "\n\n"
	for k, action := range section.Actions {
		text += "/" + strconv.Itoa(k+1) + " " + action.Label + "\n"
	}

	fsm.log.Debug("Sent message")

	_, err = fsm.client.MessagesSend(int(obj.Message.PeerID), text, params)
	if err != nil {
		panic(err)
	}
}

func paramsRandomID() url.Values {
	params := url.Values{}

	rand.Seed(time.Now().UnixNano())
	params.Set("random_id", strconv.FormatInt(rand.Int63(), 10))
	return params
}

func (fsm *Fantasm) storageSet(userID int64, key string, value string) error {
	params := url.Values{}
	params.Set("key", key)
	params.Set("value", value)
	params.Set("user_id", strconv.FormatInt(userID, 10))

	_, err := fsm.client.MakeRequest("storage.set", params)
	return err
}

func (fsm *Fantasm) storageGet(userID int64, key string) (string, error) {
	params := url.Values{}
	params.Set("key", key)
	params.Set("user_id", strconv.FormatInt(userID, 10))

	resp, err := fsm.client.MakeRequest("storage.get", params)
	if err != nil {
		return "", err
	}

	var result string

	err = json.Unmarshal(resp.Response, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}
