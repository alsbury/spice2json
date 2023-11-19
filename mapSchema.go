package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authzed/spicedb/pkg/namespace"
	"github.com/authzed/spicedb/pkg/proto/core/v1"
	"github.com/authzed/spicedb/pkg/proto/impl/v1"
)

func mapDefinition(def *corev1.NamespaceDefinition) (*Definition, error) {
	var relations []*Relation
	var permissions []*Permission
	for _, r := range def.Relation {
		kind := namespace.GetRelationKind(r)
		if kind == implv1.RelationMetadata_PERMISSION {
			permissions = append(permissions, mapPermission(r))
		} else if kind == implv1.RelationMetadata_RELATION {
			relations = append(relations, mapRelation(r))
		} else {
			return nil, fmt.Errorf("unexpected relation %q, neither permission nor relation", r.Name)
		}
	}

	splits := strings.SplitN(def.Name, "/", 2)
	var name string
	var ns string
	if len(splits) == 2 {
		ns = splits[0]
		name = splits[1]
	} else {
		name = splits[0]
		ns = ""
	}

	return &Definition{
		Name:        name,
		Namespace:   ns,
		Relations:   relations,
		Permissions: permissions,
		Comment:     getMetadataComments(def.GetMetadata()),
	}, nil
}

func mapRelation(relation *corev1.Relation) *Relation {
	var types []*RelationType
	for _, t := range relation.TypeInformation.AllowedDirectRelations {
		types = append(types, mapRelationType(t))
	}

	return &Relation{
		Name:    relation.Name,
		Comment: getMetadataComments(relation.GetMetadata()),
		Types:   types,
	}
}

func mapPermission(relation *corev1.Relation) *Permission {
	return &Permission{
		Name:    relation.Name,
		UserSet: mapUserSet(relation.GetUsersetRewrite()),
		Comment: getMetadataComments(relation.GetMetadata()),
	}
}

func mapUserSet(userset *corev1.UsersetRewrite) *UserSet {
	union := userset.GetUnion()
	if union != nil {
		return &UserSet{
			Operation: "union",
			Children:  mapUserSetChild(union.GetChild()),
		}
	}

	intersection := userset.GetIntersection()
	if intersection != nil {
		return &UserSet{
			Operation: "intersection",
			Children:  mapUserSetChild(intersection.GetChild()),
		}
	}

	exclusion := userset.GetExclusion()
	if exclusion != nil {
		return &UserSet{
			Operation: "exclusion",
			Children:  mapUserSetChild(exclusion.GetChild()),
		}
	}

	return nil
}

func mapUserSetChild(children []*corev1.SetOperation_Child) []*UserSet {
	var sets []*UserSet
	for _, child := range children {
		computed := child.GetComputedUserset()
		if computed != nil {
			sets = append(sets, &UserSet{
				Relation: computed.Relation,
			})
		}

		tuple := child.GetTupleToUserset()
		if tuple != nil {
			sets = append(sets, &UserSet{
				Relation:   tuple.Tupleset.Relation,
				Permission: tuple.ComputedUserset.Relation,
			})
		}

		set := child.GetUsersetRewrite()
		if set != nil {
			sets = append(sets, mapUserSet(set))
		}
	}
	return sets
}

func mapRelationType(relationType *corev1.AllowedRelation) *RelationType {
	Relation, ok := relationType.RelationOrWildcard.(*corev1.AllowedRelation_Relation)
	var relationName string
	if !ok {
		relationName = "*"
	} else {
		relationName = Relation.Relation
		if relationName == "..." {
			relationName = ""
		}
	}

	caveat := relationType.RequiredCaveat
	var caveatName string
	if caveat != nil {
		caveatName = caveat.CaveatName
	} else {
		caveatName = ""
	}
	return &RelationType{
		Type:     relationType.Namespace,
		Relation: relationName,
		Caveat:   caveatName,
	}
}

var commentRegex = regexp.MustCompile("(/[*]{1,2} ?|// ?| ?[*] | ?[*]?/)")

func getMetadataComments(metaData *corev1.Metadata) string {
	comment := ""
	for _, d := range metaData.GetMetadataMessage() {
		if d.GetTypeUrl() == "type.googleapis.com/impl.iv1.DocComment" {
			comment += commentRegex.ReplaceAllString(string(d.GetValue()[2:]), "") + "\n"
		}
	}
	return strings.TrimSpace(comment)
}

func mapCaveat(caveat *corev1.CaveatDefinition) *Caveat {
	var parameters []string
	for _, t := range caveat.ParameterTypes {
		parameters = append(parameters, t.TypeName)
	}

	return &Caveat{
		Name:       caveat.Name,
		Parameters: parameters,
		Comment:    getMetadataComments(caveat.Metadata),
	}
}

type Definition struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace,omitempty"`
	Relations   []*Relation   `json:"relations,omitempty"`
	Permissions []*Permission `json:"permissions,omitempty"`
	Comment     string        `json:"comment,omitempty"`
}

type Relation struct {
	Name    string          `json:"name"`
	Types   []*RelationType `json:"types"`
	Comment string          `json:"comment,omitempty"`
}

type RelationType struct {
	Type     string `json:"type"`
	Relation string `json:"relation,omitempty"`
	Caveat   string `json:"caveat,omitempty"`
}

type Permission struct {
	Name    string   `json:"name"`
	UserSet *UserSet `json:"userSet"`
	Comment string   `json:"comment,omitempty"`
}

type UserSet struct {
	Operation  string     `json:"operation,omitempty"`
	Relation   string     `json:"relation,omitempty"`
	Permission string     `json:"permission,omitempty"`
	Children   []*UserSet `json:"children,omitempty"`
}

type Caveat struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters"`
	Comment    string   `json:"comment,omitempty"`
}

type Schema struct {
	Definitions []*Definition `json:"definitions"`
	Caveats     []*Caveat     `json:"caveats,omitempty"`
}
