package chart

import "github.com/WindowsSov8forUs/sonolus-core-go/core/resource"

type Intermediate struct {
	Archetype resource.EngineArchetypeName
	Data      map[resource.EngineArchetypeDataName]any
	Sim       bool
}

const (
	archetypeTapNote                   resource.EngineArchetypeName     = "TapNote"
	archetypeSkillNote                 resource.EngineArchetypeName     = "SkillNote"
	archetypeFlickNote                 resource.EngineArchetypeName     = "FlickNote"
	archetypeDirectionalFlickNote      resource.EngineArchetypeName     = "DirectionalFlickNote"
	archetypeStraightSlideConnector    resource.EngineArchetypeName     = "StraightSlideConnector"
	archetypeCurvedSlideConnector      resource.EngineArchetypeName     = "CurvedSlideConnector"
	archetypeSlideStartNote            resource.EngineArchetypeName     = "SlideStartNote"
	archetypeSlideStartSkillNote       resource.EngineArchetypeName     = "SlideStartSkillNote"
	archetypeSlideStartFlickNote       resource.EngineArchetypeName     = "SlideStartFlickNote"
	archetypeSlideStartDirectionalNote resource.EngineArchetypeName     = "SlideStartDirectionalNote"
	archetypeSlideEndNote              resource.EngineArchetypeName     = "SlideEndNote"
	archetypeSlideEndSkillNote         resource.EngineArchetypeName     = "SlideEndSkillNote"
	archetypeSlideEndFlickNote         resource.EngineArchetypeName     = "SlideEndFlickNote"
	archetypeSlideEndDirectionalNote   resource.EngineArchetypeName     = "SlideEndDirectionalNote"
	archetypeIgnoredNote               resource.EngineArchetypeName     = "IgnoredNote"
	archetypeSlideTickNote             resource.EngineArchetypeName     = "SlideTickNote"
	archetypeSlideTickFlickNote        resource.EngineArchetypeName     = "SlideTickFlickNote"
	archetypeSlideTickDirectionalNote  resource.EngineArchetypeName     = "SlideTickDirectionalNote"
	archetypeInitialization            resource.EngineArchetypeName     = "Initialization"
	archetypeStage                     resource.EngineArchetypeName     = "Stage"
	archetypeSimLine                   resource.EngineArchetypeName     = "SimLine"
	archetypeScrollVelocity            resource.EngineArchetypeName     = "ScrollVelocity"
	dataNameLane                       resource.EngineArchetypeDataName = "lane"
	dataNameDirection                  resource.EngineArchetypeDataName = "direction"
	dataNameSize                       resource.EngineArchetypeDataName = "size"
	dataNameGroup                      resource.EngineArchetypeDataName = "group"
	dataNameValue                      resource.EngineArchetypeDataName = "value"
	dataNameFirst                      resource.EngineArchetypeDataName = "first"
	dataNamePrev                       resource.EngineArchetypeDataName = "prev"
	dataNameLong                       resource.EngineArchetypeDataName = "long"
	dataNameLast                       resource.EngineArchetypeDataName = "last"
	dataNameStart                      resource.EngineArchetypeDataName = "start"
	dataNameHead                       resource.EngineArchetypeDataName = "head"
	dataNameTail                       resource.EngineArchetypeDataName = "tail"
	dataNameEnd                        resource.EngineArchetypeDataName = "end"
	dataNameMode                       resource.EngineArchetypeDataName = "mode"
	dataNameHeadDirection              resource.EngineArchetypeDataName = "headDirection"
	dataNameHeadSize                   resource.EngineArchetypeDataName = "headSize"
	dataNameA                          resource.EngineArchetypeDataName = "a"
	dataNameB                          resource.EngineArchetypeDataName = "b"
)
