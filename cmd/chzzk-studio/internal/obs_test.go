package internal

import (
	"testing"

	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
)

func TestNewOBSClient(t *testing.T) {
	client, err := NewOBSClient("4455", "fwZkU32mADl40Cei")
	if err != nil {
		t.Fatalf("Failed to create OBS client: %v", err)
	}
	if client == nil {
		t.Fatal("OBS client is nil")
	}
}

func TestOBSClientFetchScenes(t *testing.T) {
	client, err := NewOBSClient("4455", "fwZkU32mADl40Cei")
	if err != nil {
		t.Fatalf("Failed to create OBS client: %v", err)
	}
	if client == nil {
		t.Fatal("OBS client is nil")
	}

	scenes, err := client.Scenes.GetSceneList()
	if err != nil {
		t.Fatalf("Failed to fetch scenes: %v", err)
	}
	if len(scenes.Scenes) == 0 {
		t.Log("No scenes found")
	} else {
		t.Logf("Scenes fetched: %v", scenes.Scenes)
	}
}

func TestOBSClientFetchSources(t *testing.T) {
	client, err := NewOBSClient("4455", "fwZkU32mADl40Cei")
	if err != nil {
		t.Fatalf("Failed to create OBS client: %v", err)
	}
	if client == nil {
		t.Fatal("OBS client is nil")
	}
	sceneItemListParams := sceneitems.NewGetSceneItemListParams()
	items, err := client.SceneItems.GetSceneItemList(sceneItemListParams.WithSceneName("Scene"))
	if err != nil {
		t.Fatalf("Failed to fetch active source: %v", err)
	}
	t.Logf("Scene items fetched: %v", items.SceneItems)
}

func TestOBSClientSetScene(t *testing.T) {
	client, err := NewOBSClient("4455", "fwZkU32mADl40Cei")
	if err != nil {
		t.Fatalf("Failed to create OBS client: %v", err)
	}
	if client == nil {
		t.Fatal("OBS client is nil")
	}
	p := scenes.NewSetCurrentProgramSceneParams().WithSceneName("Scene")
	resp, err := client.Scenes.SetCurrentProgramScene(p)
	if err != nil {
		t.Fatalf("Failed to set current scene: %v", err)
	}
	t.Logf("Set current program scene response: %v", resp)
	t.Log("Current scene set successfully")
}

func TestOBSAddSource(t *testing.T) {
	client, err := NewOBSClient("4455", "fwZkU32mADl40Cei")
	if err != nil {
		t.Fatalf("Failed to create OBS client: %v", err)
	}
	if client == nil {
		t.Fatal("OBS client is nil")
	}
	// Example of adding a source to a scene
	params := sceneitems.NewCreateSceneItemParams().
		WithSceneName("Scene").
		WithSourceName("SourceName")
	resp, err := client.SceneItems.CreateSceneItem(params)
	if err != nil {
		t.Fatalf("Failed to add source to scene: %v", err)
	}
	t.Logf("Add source response: %v", resp)
}
