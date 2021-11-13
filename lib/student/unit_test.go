package student

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LinehayatTestSuite struct {
	suite.Suite
	studentRepo *StudentRepo
}

// SetupTest runs at the start of each test
func (s *LinehayatTestSuite) SetupTest() {
	s.studentRepo = NewStudentRepo()
}

// TestLinehayatServer is the entry point to the test.  Just run the whole
// test suite.
func TestLinehayatServer(t *testing.T) {
	suite.Run(t, &LinehayatTestSuite{})
}

func (s *LinehayatTestSuite) Test_RemoveByUserID() {
	student := NewStudent(nil)
	// If we add just one student, and remove that student, we should have no more students.
	s.studentRepo.Add(student)
	s.Require().Equal(len(s.studentRepo.students), 1)
	s.Require().Equal(s.studentRepo.GetStudentByUserID(student.UserID), student)

	s.studentRepo.RemoveByUserID(student.UserID)
	s.Require().Empty(s.studentRepo.students)

	// If we add many students, and remove a student, that student should no longer be in the student repo.
	students := make([]*Student, 0)
	numStudents := 100
	for i := 0; i < numStudents; i++ {
		student = NewStudent(nil)
		students = append(students, student)
		s.studentRepo.Add(student)
	}
	s.Require().Equal(len(s.studentRepo.students), numStudents)

	studentToRemove := students[rand.Intn(numStudents)]
	s.studentRepo.RemoveByUserID(studentToRemove.UserID)
	s.Require().Equal(len(s.studentRepo.students), numStudents-1)

	for _, st := range s.studentRepo.students {
		s.Require().NotEqual(st.UserID, studentToRemove.UserID)
	}

	// If we remove that student again, it should be a no-op.
	s.studentRepo.RemoveByUserID(studentToRemove.UserID)
	s.Require().Equal(len(s.studentRepo.students), numStudents-1)

	for _, st := range s.studentRepo.students {
		s.Require().NotEqual(st.UserID, studentToRemove.UserID)
	}
}
