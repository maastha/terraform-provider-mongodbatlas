package team

import (
	"context"

	"go.mongodb.org/atlas-sdk/v20240805004/admin"
)

func UpdateTeamUsers(teamsAPI admin.TeamsApi, usersAPI admin.MongoDBCloudUsersApi, existingTeamUsers *admin.PaginatedApiAppUser, newUsernames []string, orgID, teamID string) error {
	validNewUsers, err := ValidateUsernames(usersAPI, newUsernames)
	if err != nil {
		return err
	}
	usersToAdd, usersToRemove, err := GetChangesForTeamUsers(existingTeamUsers.GetResults(), validNewUsers)
	if err != nil {
		return err
	}

	// Avoid leaving the team empty with no users by first making new additions, ensuring no API validation errors

	var userToAddModels []admin.AddUserToTeam
	for i := range usersToAdd {
		userToAddModels = append(userToAddModels, admin.AddUserToTeam{Id: usersToAdd[i]})
	}
	// save all users to add
	if len(userToAddModels) > 0 {
		_, _, err = teamsAPI.AddTeamUser(context.Background(), orgID, teamID, &userToAddModels).Execute()
		if err != nil {
			return err
		}
	}

	for i := range usersToRemove {
		// remove user from team
		_, err := teamsAPI.RemoveTeamUser(context.Background(), orgID, teamID, usersToRemove[i]).Execute()
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateUsernames(c admin.MongoDBCloudUsersApi, usernames []string) ([]admin.CloudAppUser, error) {
	var validUsers []admin.CloudAppUser
	for _, elem := range usernames {
		userToAdd, _, err := c.GetUserByUsername(context.Background(), elem).Execute()
		if err != nil {
			return nil, err
		}
		validUsers = append(validUsers, *userToAdd)
	}
	return validUsers, nil
}

func GetChangesForTeamUsers(currentUsers, newUsers []admin.CloudAppUser) (toAdd, toDelete []string, err error) {
	// Create two sets to store the elements in A and B
	currentUsersSet := InitUserSet(currentUsers)
	newUsersSet := InitUserSet(newUsers)

	// Iterate over the elements in B and add them to the toAdd array if they are not in A
	for elem := range newUsersSet {
		if _, ok := currentUsersSet[elem]; !ok {
			toAdd = append(toAdd, elem)
		}
	}

	// Iterate over the elements in A and add them to the toDelete array if they are not in B
	for elem := range currentUsersSet {
		if _, ok := newUsersSet[elem]; !ok {
			toDelete = append(toDelete, elem)
		}
	}

	return toAdd, toDelete, nil
}

func InitUserSet(users []admin.CloudAppUser) map[string]interface{} {
	usersSet := make(map[string]interface{}, len(users))
	for i := 0; i < len(users); i++ {
		usersSet[users[i].GetId()] = true
	}
	return usersSet
}