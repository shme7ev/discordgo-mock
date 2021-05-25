package mockrest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
)

func (roundTripper *RoundTripper) addHandlersGuilds(apiVersion string) {
	pathGuildID := fmt.Sprintf("%s/%s/%s", apiVersion, resourceGuilds, resourceGuildID)

	subrouter := roundTripper.router.PathPrefix(pathGuildID).Subrouter()

	pathRoles := fmt.Sprintf("/%s", resourceRoles)
	pathRoleID := fmt.Sprintf("%s/%s", pathRoles, resourceRoleID)
	pathMembers := fmt.Sprintf("/%s", resourceMembers)
	pathMembersUserID := fmt.Sprintf("%s/%s", pathMembers, resourceUserID)
	pathMembersUserIDRoles := fmt.Sprintf("%s/%s", pathMembersUserID, resourceRoles)
	pathMembersUserIDRoleID := fmt.Sprintf("%s/%s", pathMembersUserIDRoles, resourceRoleID)

	getHandlers := subrouter.Methods(http.MethodGet).Subrouter()
	getHandlers.HandleFunc("", roundTripper.guildGET)
	getHandlers.HandleFunc(pathRoles, roundTripper.guildRolesGET)
	getHandlers.HandleFunc(pathRoleID, roundTripper.guildRolesGET)
	getHandlers.HandleFunc(pathMembers, roundTripper.guildMembersGET)
	getHandlers.HandleFunc(pathMembersUserID, roundTripper.guildMembersGET)
	getHandlers.HandleFunc(pathMembersUserIDRoles, roundTripper.guildMembersGET)
	getHandlers.HandleFunc(pathMembersUserIDRoleID, roundTripper.guildMembersGET)

	postHandlers := subrouter.Methods(http.MethodPost).Subrouter()
	postHandlers.HandleFunc(pathRoles, roundTripper.guildRolesPOST)
	postHandlers.HandleFunc(pathMembers, roundTripper.guildMembersPOST)

	putHandlers := subrouter.Methods(http.MethodPut).Subrouter()
	putHandlers.HandleFunc(pathMembersUserIDRoleID, roundTripper.guildMemberRolesPUT)

	patchHandlers := subrouter.Methods(http.MethodPatch).Subrouter()
	patchHandlers.HandleFunc(pathRoleID, roundTripper.guildRolesPATCH)
	patchHandlers.HandleFunc(pathMembersUserID, roundTripper.guildMembersPATCH)

	deleteHandlers := subrouter.Methods(http.MethodDelete).Subrouter()
	deleteHandlers.HandleFunc(pathRoleID, roundTripper.guildRolesDELETE)
	deleteHandlers.HandleFunc(pathMembersUserID, roundTripper.guildMembersDELETE)
	deleteHandlers.HandleFunc(pathMembersUserIDRoleID, roundTripper.guildMemberRolesDELETE)
}

func (roundTripper *RoundTripper) guildGET(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID, foundGuildID := vars[resourceGuildIDKey]

	var respBody []byte

	switch {
	case foundGuildID:
		guild, err := roundTripper.state.Guild(guildID)
		if err != nil {
			sendError(w, err)
			return
		}

		respBody, err = json.Marshal(guild)
		if err != nil {
			sendError(w, err)
			return
		}
	default:
		var err error

		respBody, err = json.Marshal(roundTripper.state.Guilds)
		if err != nil {
			sendError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(respBody)
}

func (roundTripper *RoundTripper) guildRolesGET(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	roleID, foundRoleID := vars[resourceRoleIDKey]

	var respBody []byte

	switch {
	case foundRoleID:
		role, err := roundTripper.state.Role(guildID, roleID)
		if err != nil {
			sendError(w, err)
			return
		}

		respBody, err = json.Marshal(role)
		if err != nil {
			sendError(w, err)
			return
		}
	default:
		guild, err := roundTripper.state.Guild(guildID)
		if err != nil {
			sendError(w, err)
			return
		}

		respBody, err = json.Marshal(guild.Roles)
		if err != nil {
			sendError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(respBody)
}

func (roundTripper *RoundTripper) guildMembersGET(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	userID, foundUserID := vars[resourceUserIDKey]

	var respBody []byte

	switch {
	case foundUserID:
		member, err := roundTripper.state.Member(guildID, userID)
		if err != nil {
			sendError(w, err)
			return
		}

		respBody, err = json.Marshal(member)
		if err != nil {
			sendError(w, err)
			return
		}
	default:
		guild, err := roundTripper.state.Guild(guildID)
		if err != nil {
			sendError(w, err)
			return
		}

		respBody, err = json.Marshal(guild.Members)
		if err != nil {
			sendError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(respBody)
}

func (roundTripper *RoundTripper) guildRolesPOST(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]

	role := &discordgo.Role{
		ID: randString(),
	}

	err := roundTripper.state.RoleAdd(guildID, role)
	if err != nil {
		sendError(w, fmt.Errorf("error adding role to state: %w", err))
		return
	}

	respBody, err := json.Marshal(role)
	if err != nil {
		sendError(w, fmt.Errorf("error marshaling role to response body: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(respBody)
}

func (roundTripper *RoundTripper) guildMembersPOST(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]

	guild, err := roundTripper.state.Guild(guildID)
	if err != nil {
		sendError(w, err)
		return
	}

	member := &discordgo.Member{
		GuildID: guild.ID,
	}

	err = roundTripper.state.MemberAdd(member)
	if err != nil {
		sendError(w, err)
		return
	}

	guild.MemberCount++

	respBody, err := json.Marshal(member)
	if err != nil {
		sendError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(respBody)
}

func (roundTripper *RoundTripper) guildMemberRolesPUT(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	userID := vars[resourceUserIDKey]
	roleID := vars[resourceRoleIDKey]

	member, err := roundTripper.state.Member(guildID, userID)
	if err != nil {
		sendError(w, fmt.Errorf("member not found: %w", err))
		return
	}

	member.Roles = append(member.Roles, roleID)

	err = roundTripper.state.MemberAdd(member)
	if err != nil {
		sendError(w, fmt.Errorf("unable to add or update member: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte{})
}

func (roundTripper *RoundTripper) guildRolesPATCH(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	roleID := vars[resourceRoleIDKey]

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, err)
		return
	}

	defer func() {
		_ = r.Body.Close()
	}()

	role, err := roundTripper.state.Role(guildID, roleID)
	if err != nil {
		sendError(w, err)
		return
	}

	err = json.Unmarshal(reqBody, role)
	if err != nil {
		sendError(w, err)
		return
	}

	err = roundTripper.state.RoleAdd(guildID, role)
	if err != nil {
		sendError(w, err)
		return
	}

	respBody, err := json.Marshal(role)
	if err != nil {
		sendError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(respBody)
}

func (roundTripper *RoundTripper) guildMembersPATCH(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	userID := vars[resourceUserIDKey]

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, err)
		return
	}

	defer func() {
		_ = r.Body.Close()
	}()

	member, err := roundTripper.state.Member(guildID, userID)
	if err != nil {
		sendError(w, fmt.Errorf("member not found: %w", err))
		return
	}

	err = json.Unmarshal(reqBody, member)
	if err != nil {
		sendError(w, err)
		return
	}

	err = roundTripper.state.MemberAdd(member)
	if err != nil {
		sendError(w, fmt.Errorf("unable to update member: %w", err))
		return
	}

	respBody, err := json.Marshal(member)
	if err != nil {
		sendError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(respBody)
}

func (roundTripper *RoundTripper) guildRolesDELETE(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	roleID := vars[resourceRoleIDKey]

	err := roundTripper.state.RoleRemove(guildID, roleID)
	if err != nil {
		sendError(w, fmt.Errorf("unable to remove role: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte{})
}

func (roundTripper *RoundTripper) guildMembersDELETE(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	userID := vars[resourceUserIDKey]

	member, err := roundTripper.state.Member(guildID, userID)
	if err != nil {
		sendError(w, fmt.Errorf("member not found: %w", err))
		return
	}

	err = roundTripper.state.MemberRemove(member)
	if err != nil {
		sendError(w, fmt.Errorf("unable to remove member: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte{})
}

func (roundTripper *RoundTripper) guildMemberRolesDELETE(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars[resourceGuildIDKey]
	userID := vars[resourceUserIDKey]
	roleID := vars[resourceRoleIDKey]

	member, err := roundTripper.state.Member(guildID, userID)
	if err != nil {
		sendError(w, fmt.Errorf("member not found: %w", err))
		return
	}

	index := -1

	for i, memberRoleID := range member.Roles {
		if memberRoleID == roleID {
			index = i
			break
		}
	}

	if index != -1 {
		member.Roles = append(member.Roles[:index], member.Roles[index:len(member.Roles)]...)
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte{})
}