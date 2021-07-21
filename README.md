Client example (for testing)

# Array
Students Array

OnSetStatus -> Notify Volunteers
OnSetStatus -> Notify connected

Volunteers Array

OnSetStatus -> Notify Volunteers
OnSetStatus -> Notify connected

# Cases to handle
## Student becomes disconnected

## Student ends conversation

## Student request for chat

## Volunteer becomes disconnected

## Volunteer ends conversation

## Volunteer accepts Student Chat

## Volunteer logs in

## Volunteer logs out

# Possible states
## Student
* Wait
* Chat (Disconnect)
* Chat (Active)

## Volunteer
* Free
* Chat (Disconnect)
* Chat (Active)

# States Entry (Student)
Wait -> Chat (Active)
Chat (Active) -> Chat (Disconnect)
Chat (Disconnect) -> Chat (Active)
Chat (Active) -> Deleted 

# States Entry (Volunteer)
Free -> Chat (Active)
Chat (Active) -> Chat (Disconnect)
Chat (Active) -> Free
Chat (Disconnect) -> Free
Chat (Disconnect) -> Chat (Active)

# ConnectedList
A list listing pairs of connected people

A single source of truth for checking who is connected to who

* map[Student][Volunteer]
    * ConnectedToVolunteer(Student)
    * ConnectedToStudent(Volunteer)
    * NotifyConnectedToVolunteer(Student, message)
    * NotifyConnectedToStudent(Volunteer, message)

# When States Entry occur (Student)
* Wait -> Chat(Active)
    * When Volunteer accepts student request
    * **Action**
        * Tell Student that chat request is accepted.
        * Add new ConnectedTo entry to ConnectedList
        * Notify all volunteers that student's request has been accepted
* Chat(Active) -> Chat(Disconnect)
    * When Student socket disconnects (error on read)
    * **Action**
        * To differentiate between Refresh vs Disconnect
            * Set a Refresh timeout
                * After pass refresh timeout, notify volunteer that student has disconnect
        * When disconnect (after pass Refresh timeout)
            * Tell connectedTo volunteer that student has suddenly disconnect
                * If volunteer is disconnected, skip telling.
        * Close socket connection and set to Nil.
* Chat(Disconnect) -> Chat(Active)
    * When Student socket connects with a RECONNECT message and a UserID.
    * **Action**
        * Tell connectedTo volunteer that student has suddenly reconnect
            * If volunteer is disconnected, skip telling.
* Chat(Active) -> Deleted 
    * When Student ends conversation
    * **Action**
        * Tell volunteer that student has ended conversation
            * If volunteer is disconnected, skip telling.
        * Delete from ConnectedList
* Chat(Disconnected) -> Deleted
    * When Student time out
    * **Action**
        * Tell volunteer that student has permanently disconnected (exceed disconnect timeout)
        * Delete from ConnectedList
        

# When states entry occur (Volunteer)
* Free -> Chat (Active)
    * When Volunteer accepts student request
        * **Action**
            * Tell Student that chat request is accepted.
            * Add new ConnectedTo entry to ConnectedList
            * Notify all volunteers that student's request has been accepted
            * Notify the login volunteer on all volunteers & waiting students status

* Chat (Active) -> Chat (Disconnect)
    * When Volunteer socket disconnects (error on read)
    * **Action**
        * To differentiate between Refresh vs Disconnect
            * Set a Refresh timeout
                * After pass refresh timeout, notify student that volunteer has disconnect
        * When disconnect (after pass Refresh timeout)
            * Tell connectedTo student that volunteer has suddenly disconnect
                * If student is disconnected, skip telling.
        * Close volunteer's socket connection and set to Nil.

* Chat (Active) -> Free
    * When either Volunteer or ConnectedTo Student ends conversation
    * **Action**
        * Notify all volunteers
        * If Volunteer ends conversation, Notify student
        * If Student ends conversation, Notify volunteer
        * If Socket is nil, skip notifying
        * Delete from ConnectedList

* Chat (Disconnect) -> Offline(Deleted)
    * When volunteer passes the Disconnect timeout duration
    * **Action**
        * Notify connected Student
        * If socket is nil, skip notifying

* Chat (Disconnect) -> Chat (Active)
    * When volunteer reconnects
    * **Action**
        * Notify connected student
        * Notify all volunteers


## Note
* Every change in Volunteer States result in updates to all volunteers.

* Every change in Student waiting state result in updates to all volunteers.

* Every change in Student chat state (active -> disconnect, disconnect -> active) doesn't result in updates to all volunteers, but only to the connected volunteer.

* When Student state changes to offline, ConnectedTo volunteer state will also change from Active/Disconnect -> Free. This will update all volunteers.

## Danger
When notifying all volunteers, some volunteers could potentially in disconnect.
    * Skip notifying these volunteers.

When volunteer reconnects without passing timeout duration, rese

## Edge cases
* Student reconnects but already time out
    * Use GET endpoint to handle, sends a request to server to reconnect
    * If can reconnect & Volunteer Disconnects or Active, setup a RECONNECT socket connection
    * If can reconnect & Volunteer ends conversation, tell Student to get in waiting list again because Volunteer already ended conversation
    * If cannot reconnect (timeout), tell Student to get in waiting list again

* How to handle, when volunteer disconnects when free?
    * Socket would error on read.
    * If we emit "disconnect" event (normally, as we handle disconnects when chat), we need to:
        * notify the connected student
            * If cannot be found from connected list, we simply skip
        * set socket to null
        * notify all volunteers, then wait timeout
    * ✅ Or we could handle it separately by deleting the volunteer from database
        * `if v.fsm.Is("free") socket.Close; volunteers.Remove("email");`


# Endpoints
* GET: Volunteer permission to login

