package notes

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Note struct {
	ID       uuid.UUID `json:"id" xml:"id,attr" bson:"_id"`
	Folder   string    `json:"path" xml:"folder,attr,omitempty" storm:"index"`
	Title    string    `json:"title" xml:"Meta>Title"`
	Owner    uuid.UUID `json:"owner" xml:"Meta>Owner" storm:"index"`
	Created  time.Time `json:"created" xml:"Meta>Created"`
	Modified time.Time `json:"modified" xml:"Meta>Modified" storm:"index"`
	Tags     []string  `json:"tags" xml:"Meta>Tags>Tag,omitempty" storm:"index"`
	Body     string    `json:"body"`
}
