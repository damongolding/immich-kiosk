package immich

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
)

// convertFaceResponse takes a slice of AssetFaceResponse and converts it into a slice of Person.
// It groups faces by person ID, ensuring each person appears only once in the result with all
// their associated faces grouped together. The function creates Face objects from the response
// data and either adds them to existing Person entries or creates new Person entries as needed.
func convertFaceResponse(faceResponse []AssetFaceResponse) []Person {
	personMap := make(map[string]*Person)

	for _, face := range faceResponse {
		personID := face.Person.ID

		// Create Face object
		newFace := Face{
			ID:            face.ID,
			ImageHeight:   face.ImageHeight,
			ImageWidth:    face.ImageWidth,
			BoundingBoxX1: face.BoundingBoxX1,
			BoundingBoxX2: face.BoundingBoxX2,
			BoundingBoxY1: face.BoundingBoxY1,
			BoundingBoxY2: face.BoundingBoxY2,
		}

		// Check if we've already seen this person
		if person, exists := personMap[personID]; exists {
			// Add face to existing person
			person.Faces = append(person.Faces, newFace)
		} else {
			// Create new person with their first face
			personMap[personID] = &Person{
				ID:            face.Person.ID,
				Name:          face.Person.Name,
				BirthDate:     face.Person.BirthDate,
				ThumbnailPath: face.Person.ThumbnailPath,
				IsHidden:      face.Person.IsHidden,
				Faces:         []Face{newFace},
			}
		}
	}

	// Convert map to slice if needed
	people := make([]Person, 0, len(personMap))
	for _, person := range personMap {
		people = append(people, *person)
	}

	return people
}

// CheckForFaces queries the Immich API to detect faces in the asset and adds them
// to the asset's People slice. It takes requestID and deviceID parameters for API
// call tracking. The function handles URL parsing, making the API request, and
// unmarshaling the response into Face structs. Any errors are logged and will
// abort the operation.
func (a *Asset) CheckForFaces(requestID, deviceID string) {

	var faceResponse []AssetFaceResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(faceResponse, err, nil, "")
		log.Error("parsing faces url", "err", err)
		return
	}

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/faces",
		RawQuery: "id=" + a.ID,
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, faceResponse)
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		_, _, err = immichAPIFail(faceResponse, err, body, apiURL.String())
		log.Error("adding faces", "err", err)
		return
	}

	err = json.Unmarshal(body, &faceResponse)
	if err != nil {
		_, _, err = immichAPIFail(faceResponse, err, body, apiURL.String())
		log.Error("adding faces", "err", err)
		return
	}

	people := convertFaceResponse(faceResponse)

	a.People = people
}
