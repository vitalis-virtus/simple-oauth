package models

type MainProfileInfo struct {
	LocalizedLastName  string `json: "localizedLastName"`
	LocalizedFirstName string `json: "localizedFirstName`
	ID                 string `json: "id`
}

type ProfilePictureInfo struct {
	Elements []string `njson: "profilePicture.displayImage~.elements.#.identifiers.#.identifier"`
}
