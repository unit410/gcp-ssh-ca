package ca

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/cloudresourcemanager/v1"
	cloudresourcemanagerv2 "google.golang.org/api/cloudresourcemanager/v2"
)

// get active Project IDs within all provided Folder IDs
func getActiveProjectIDs(folderIDs []string) []string {
	projectIDs := []string{}

	// Search for all project IDs
	for _, folderID := range folderIDs {
		subFoldersIDs := getAllSubFoldersIDs(folderID)
		for _, subFolderID := range subFoldersIDs {
			projectIDs = append(projectIDs, getActiveProjectsIDsInFolder(subFolderID)...)
		}
	}

	return projectIDs
}

// get all active Project IDs directly within the provided GCP Folder ID
func getActiveProjectsIDsInFolder(folderID string) []string {
	ctx := context.Background()
	cloudresourcemanagerService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	projectListCall := cloudresourcemanagerService.Projects.List()
	queryFilter := fmt.Sprintf("parent.type:folder parent.id:%v", folderID)
	projectListCall.Filter(queryFilter)
	response, err := projectListCall.Do()
	if err != nil {
		log.Fatal(err)
	}
	projectIds := []string{}
	for _, p := range response.Projects {
		if p.LifecycleState == "ACTIVE" {
			projectIds = append(projectIds, p.ProjectId)
		}
	}
	return projectIds
}

// getAllSubFoldersIDs returns a slice of the provided parentFolderID
// and any/all subfolder IDs.
func getAllSubFoldersIDs(parentFolderID string) []string {
	return []string{parentFolderID}

	// TODO: There's a bug in the Google SDK that prevents this from working.
	//       wait until that's fixed before using this ¯\_(ツ)_/¯
	//       https://github.com/terraform-providers/terraform-provider-google/issues/4276
	ctx := context.Background()
	cloudresourcemanagerService, err := cloudresourcemanagerv2.NewService(ctx)
	folderListCall := cloudresourcemanagerService.Folders.List()
	folderListCall.Parent(parentFolderID)
	response, err := folderListCall.Do()
	if err != nil {
		log.Fatal(err)
	}

	allFolders := []string{parentFolderID}
	for _, folder := range response.Folders {
		allFolders = append(allFolders, getAllSubFoldersIDs(folder.Name)...)
	}
	return allFolders
}
