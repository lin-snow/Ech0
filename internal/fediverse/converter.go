package fediverse

import (
	"encoding/json"
	"fmt"
	"time"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	model "github.com/lin-snow/ech0/internal/model/fediverse"
	fileUtil "github.com/lin-snow/ech0/internal/util/file"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	mdUtil "github.com/lin-snow/ech0/internal/util/md"
)

//==============================================================================
//	Convert
//==============================================================================

// ConvertEchoToActivity 将 Echo 转换为 ActivityPub Activity
func (core *FediverseCore) ConvertEchoToActivity(
	echo *echoModel.Echo,
	actor *model.Actor,
	serverURL string,
) model.Activity {
	obj := core.ConvertEchoToObject(echo, actor, serverURL)

	activityID := fmt.Sprintf("%s/activities/%d", serverURL, echo.ID)

	activity := model.Activity{
		Context: []any{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		ActivityID: activityID,
		Type:       model.ActivityTypeCreate,
		ActorID:    actor.ID,
		ActorURL:   actor.ID,
		Object:     obj.ObjectID,
		ObjectID:   obj.ObjectID,
		ObjectType: obj.Type,
		Published:  echo.CreatedAt,
		To:         obj.To,
		Cc:         []string{actor.Followers},
		Summary:    "",
		Delivered:  false,
		CreatedAt:  time.Now().UTC(),
	}

	activityJSON, _ := json.Marshal(activity)
	activity.ActivityJSON = string(activityJSON)
	return activity
}

// ConvertEchoToObject 将 Echo 转换为 ActivityPub Object
func (core *FediverseCore) ConvertEchoToObject(
	echo *echoModel.Echo,
	actor *model.Actor,
	serverURL string,
) model.Object {
	var attachments []model.Attachment
	for i := range echo.Images {
		attachments = append(attachments, model.Attachment{
			Type:      "Image",
			MediaType: httpUtil.GetMIMETypeFromFilenameOrURL(echo.Images[i].ImageURL),
			URL:       fileUtil.GetImageURL(echo.Images[i], serverURL),
		})
	}

	return model.Object{
		Context: []any{
			"https://www.w3.org/ns/activitystreams",
		},
		ObjectID: fmt.Sprintf("%s/objects/%d", serverURL, echo.ID),
		Type:     "Note",
		Content:  string(mdUtil.MdToHTML([]byte(echo.Content))),
		Source: map[string]any{
			"mediaType": "text/markdown",
			"content":   echo.Content,
		},
		AttributedTo: actor.ID,
		Published:    echo.CreatedAt,
		To: []string{
			"https://www.w3.org/ns/activitystreams#Public",
		},
		Attachments: attachments,
	}
}
