package main

import (
	"fmt"
	"net/http"
	"rest-go-gin/internal/database"
	"strconv"

	"github.com/gin-gonic/gin"
)

func(app *application) createEvent(c *gin.Context) {
	var event database.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
	}

	user := app.getUserFromContext(c)
	event.OwnerId = user.Id

	err := app.models.Events.Insert(&event)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// getEvents return all events
// 
// @Summary Returns all events
// @Description Returns all events
// @Tags Events
// @Accept Json
// @Produce json
// @Success 200 {object} []database.Event
// @Router /api/v1/events [get]
func (app *application) getAllEvents(c *gin.Context){
	events, err := app.models.Events.GetAll()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve events"})
	}

	c.JSON(http.StatusOK, events)
}

func (app *application) getEvent(c *gin.Context){
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid event ID"})
	}
	event, err := app.models.Events.Get(id)

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error":"Event not found"})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve event"})
	}

	c.JSON(http.StatusOK, event)
}

func (app *application) updateEvent(c *gin.Context){
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid event ID"})
	}

	user := app.getUserFromContext(c)
	existingEvent, err := app.models.Events.Get(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve event"})
		return
	}

	if existingEvent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error":"Event not found"})
		return
	}

	if existingEvent.OwnerId != user.Id{
		c.JSON(http.StatusForbidden, gin.H{"error":"You are not authorized to update this event"})
		return
	}

	updatedEvent := &database.Event{}

	if err := c.ShouldBindJSON(updatedEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}

	

	updatedEvent.Id = id

	if err := app.models.Events.Update(updatedEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to update event"})
		return
	}
	c.JSON(http.StatusOK, updatedEvent)
}

func (app *application) deleteEvent(c *gin.Context){
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid event ID"})
	}

	user := app.getUserFromContext(c)
	existingEvent, err := app.models.Events.Get(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve event"})
		return
	}

	if existingEvent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error":"Event not found"})
		return
	}
	if existingEvent.OwnerId != user.Id{
		c.JSON(http.StatusForbidden, gin.H{"error":"You are not authorized to delete this event"})
		return
	}

	if err := app.models.Events.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to delete event"})
	}

	c.JSON(http.StatusNoContent, nil)
}

func (app *application) addAttendeeToEvent(c *gin.Context){
	eventId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid event Id"})
		return
	}

	userId, err := strconv.Atoi(c.Param("userId"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid user Id"})
		return
	}

	event, err := app.models.Events.Get(eventId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve event"})
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error":"Event not found"})
	}

	userToAdd, err := app.models.Users.Get(userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve user"})
		return
	}
	if userToAdd == nil {
		c.JSON(http.StatusNotFound, gin.H{"error":"User not found"})
	}

	user := app.getUserFromContext(c)

	if event.OwnerId != user.Id{
		c.JSON(http.StatusForbidden, gin.H{"error":"You are not authorized to add an attendee"})
		return
	}


	existingAttendee, err := app.models.Attendees.GetByEventAndAttendee(event.Id, userToAdd.Id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve attendee"})
		return
	}
	if existingAttendee != nil {
		c.JSON(http.StatusConflict, gin.H{"error":"Attendee already exists"})
	}

	attendee := database.Attendee{

		EventId: event.Id,
		UserId: userToAdd.Id,
	}

	_, err = app.models.Attendees.Insert(&attendee)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to add attendee"})
		return
	}

	c.JSON(http.StatusCreated, attendee)

}

func (app *application) getAttendeesForEvent(c *gin.Context){

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid event id"})
		return
	}

	users, err := app.models.Attendees.GetAttendeesByEvent(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to retrieve attendee"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (app *application) deleteAttendeeFromEvent( c *gin.Context){
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid event id"})
		return
	}

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid user id"})
		return
	}

	event, err := app.models.Events.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	if event == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
	}

	user := app.getUserFromContext(c)
	if event.OwnerId != user.Id {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to add content"})
	}


	err = app.models.Attendees.Delete(userId, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendee"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (app *application) getEventsByAttendee(c *gin.Context){
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid attendee id"})
		return
	}
	
	events, err := app.models.Attendees.GetEventsByAttendee(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to get events"})
		return
	}

	c.JSON(http.StatusOK, events)

}