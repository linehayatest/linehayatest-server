package student

var idGenerator userIDGenerator = newUserIDGenerator()

type userIDGenerator struct {
	currentUserID int
}

func newUserIDGenerator() userIDGenerator {
	return userIDGenerator{0}
}

func (u *userIDGenerator) getNewID() int {
	u.currentUserID += 1
	return u.currentUserID
}
