package activitystreams2

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodeImpliedActivity1(t *testing.T) {
	var f Feed
	if err := xml.Unmarshal([]byte(fixtureImpliedActivity1), &f); err != nil {
		panic(err)
	}

	assert.Equal(t, "newsstream timeline", f.Title, "Feed.Title")
	assert.Equal(t, "https://quitter.im/api/statuses/user_timeline/235.atom", f.ID, "Feed.ID")
	assert.Equal(t, "2017-04-15T04:12:30Z", f.Updated.Format(time.RFC3339), "Feed.Updated")

	author := f.Author
	assert.NotNil(t, author, "Feed.Author")
	assert.Equal(t, "", author.GetID(), "Author.GetID")
	assert.Equal(t, "newsstream", author.GetName(), "Author.GetName")
	assert.Equal(t, "Realtime News", author.GetSummary(), "Author.GetSummary")
	assert.Equal(t, "", author.GetRepresentativeImage(), "Author.GetRepresentativeImage")
	assert.Equal(t, "https://quitter.im/newsstream", author.GetPermalink(), "Author.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/person", author.GetObjectType(), "Author.GetObjectType")

	activities := f.GetActivities()
	assert.Len(t, activities, 1)

	activity := activities[0]
	assert.NotNil(t, activity)
	assert.Equal(t, "2017-04-15T04:12:24Z", activity.GetTime().Format(time.RFC3339), "Entry.GetTime")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/post", activity.GetVerb(), "Entry.GetVerb")
	assert.Equal(t, "New note by newsstream", activity.GetTitle(), "Entry.GetTitle")

	actor := activity.GetActor()
	assert.NotNil(t, actor, "Entry.GetActor")
	assert.Equal(t, "", actor.GetID(), "Author.GetID")
	assert.Equal(t, "newsstream", actor.GetName(), "Author.GetName")
	assert.Equal(t, "Realtime News", actor.GetSummary(), "Author.GetSummary")
	assert.Equal(t, "", actor.GetRepresentativeImage(), "Author.GetRepresentativeImage")
	assert.Equal(t, "https://quitter.im/newsstream", actor.GetPermalink(), "Author.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/person", actor.GetObjectType(), "Author.GetObjectType")

	object := activity.GetObject().(NoteLike)
	assert.NotNil(t, object, "Entry.GetObject")
	assert.Equal(t, "tag:quitter.im,2017-04-15:noticeId=1415216:objectType=note", object.GetID(), "Entry.GetID")
	assert.Equal(t, "New note by newsstream", object.GetName(), "Entry.GetName")
	assert.Equal(t, "", object.GetSummary(), "Entry.GetSummary")
	assert.Equal(t, "", object.GetRepresentativeImage(), "Entry.GetRepresentativeImage")
	assert.Equal(t, "https://quitter.im/notice/1415216", object.GetPermalink(), "Entry.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/note", object.GetObjectType(), "Entry.GetObjectType")

	assert.Exactly(t, activity, object)
}

func TestDecodeImpliedActivity2(t *testing.T) {
	var f Feed
	if err := xml.Unmarshal([]byte(fixtureImpliedActivity2), &f); err != nil {
		panic(err)
	}

	assert.Equal(t, "✨ Milan 🐘", f.Title, "Feed.Title")
	assert.Equal(t, "https://social.tchncs.de/users/milan.atom", f.ID, "Feed.ID")
	assert.Equal(t, "2017-04-10T21:24:16Z", f.Updated.Format(time.RFC3339), "Feed.Updated")

	activities := f.GetActivities()
	assert.Len(t, activities, 1)

	activity := activities[0]
	assert.NotNil(t, activity)
	assert.Equal(t, "2017-04-15T06:59:43Z", activity.GetTime().Format(time.RFC3339), "Entry.GetTime")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/post", activity.GetVerb(), "Entry.GetVerb")
	assert.Equal(t, "New status by milan", activity.GetTitle(), "Entry.GetTitle")

	actor := activity.GetActor()
	assert.NotNil(t, actor, "Entry.GetActor")
	assert.Equal(t, "https://social.tchncs.de/users/milan", actor.GetID(), "Author.GetID")
	assert.Equal(t, "milan", actor.GetName(), "Author.GetName")
	assert.Equal(t, "Odd humor, in love with administrating Linux servers.\nWanna help me out a bit? https://tchncs.de/donate", actor.GetSummary(), "Author.GetSummary")
	assert.Equal(t, "", actor.GetRepresentativeImage(), "Author.GetRepresentativeImage")
	assert.Equal(t, "https://social.tchncs.de/@milan", actor.GetPermalink(), "Author.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/person", actor.GetObjectType(), "Author.GetObjectType")

	object := activity.GetObject()
	assert.NotNil(t, object, "Entry.GetObject")
	assert.Equal(t, "tag:social.tchncs.de,2017-04-15:objectId=755238:objectType=Status", object.GetID(), "Entry.GetID")
	assert.Equal(t, "New status by milan", object.GetName(), "Entry.GetName")
	assert.Equal(t, "", object.GetSummary(), "Entry.GetSummary")
	assert.Equal(t, "", object.GetRepresentativeImage(), "Entry.GetRepresentativeImage")
	assert.Equal(t, "https://social.tchncs.de/users/milan/updates/45844", object.GetPermalink(), "Entry.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/note", object.GetObjectType(), "Entry.GetObjectType")

	assert.Exactly(t, activity, object)
}

func TestDecodeNestedActivity(t *testing.T) {
	var f Feed
	if err := xml.Unmarshal([]byte(fixtureNested), &f); err != nil {
		panic(err)
	}

	assert.Equal(t, "lambadalambda timeline", f.Title, "Feed.Title")
	assert.Equal(t, "https://social.heldscal.la/api/statuses/user_timeline/23211.atom", f.ID, "Feed.ID")
	assert.Equal(t, "2017-04-13T06:57:43Z", f.Updated.Format(time.RFC3339), "Feed.Updated")

	activities := f.GetActivities()
	assert.Len(t, activities, 1)

	activity := activities[0]
	assert.NotNil(t, activity)
	assert.Equal(t, "2017-04-13T06:57:43Z", activity.GetTime().Format(time.RFC3339), "Entry.GetTime")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/share", activity.GetVerb(), "Entry.GetVerb")
	assert.Equal(t, "lambadalambda repeated a notice by lohang", activity.GetTitle(), "Entry.GetTitle")
	assert.Equal(t, "", activity.GetSummary(), "Entry.GetSummary")
	assert.Equal(t, "", activity.GetObjectType(), "Entry.GetObjectType")

	actor := activity.GetActor()
	assert.NotNil(t, actor, "Entry.GetActor")
	assert.Equal(t, "", actor.GetID(), "Author.GetID")
	assert.Equal(t, "lambadalambda", actor.GetName(), "Author.GetName")
	assert.Equal(t, "Call me Deacon Blues.", actor.GetSummary(), "Author.GetSummary")
	assert.Equal(t, "", actor.GetRepresentativeImage(), "Author.GetRepresentativeImage")
	assert.Equal(t, "https://social.heldscal.la/lambadalambda", actor.GetPermalink(), "Author.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/person", actor.GetObjectType(), "Author.GetObjectType")

	object1 := activity.GetObject().(ActivityLike)
	assert.NotNil(t, object1, "Activity.GetObject")
	assert.Equal(t, "tag:gnusocial.no,2017-04-13:noticeId=1938446:objectType=note", object1.GetID(), "Activity.GetObject.GetID")
	assert.Equal(t, "lohang repeated a notice by sim", object1.GetName(), "Activity.GetObject.GetName")
	assert.Equal(t, "RT @<a href=\"https://social.heldscal.la/user/26137\" class=\"h-card mention\" title=\"Sim\">sim</a> Dear instance admins, I am the moderator for both sealion.club and freezepeach.xyz. I would like to reach out to you in the interests of compiling a few lists relating to whether our instances can communicate or not. It is in the interests of transparency, so that I can help my users by letting them know whether they can communicate with your instance or not. We don't yet have a visible way of knowing, and therefore I think it can cause confusion and frustration as I know it has for me. It has also come to my attention that knowing this helps for debugging purposes for admins, so we know whether it is a block or a bug messing things up. It saves time to know, and we can move onto making things better on our instances. I am still thinking how best to make things transparent, my intent is not to create a 'target list'... but to let my users know why they are having trouble contacting your instance if they are. That way they can decide what to do with that information, perhaps ...", object1.GetSummary(), "Activity.GetObject.GetSummary")
	assert.Equal(t, "", object1.GetRepresentativeImage(), "Activity.GetObject.GetRepresentativeImage")
	assert.Equal(t, "https://gnusocial.no/notice/1938446", object1.GetPermalink(), "Activity.GetObject.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/activity", object1.GetObjectType(), "Activity.GetObject.GetObjectType")

	object2 := object1.GetObject().(ActivityLike)
	assert.NotNil(t, object2, "Activity.GetObject")
	assert.Equal(t, "tag:social.heldscal.la,2017-04-13:noticeId=1661724:objectType=note", object2.GetID(), "Activity.GetObject.GetID")
	assert.Equal(t, "", object2.GetName(), "Activity.GetObject.GetName")
	assert.Equal(t, "Dear instance admins, I am the moderator for both sealion.club and freezepeach.xyz. I would like to reach out to you in the interests of compiling a few lists relating to whether our instances can communicate or not.<br /> <br /> It is in the interests of transparency, so that I can help my users by letting them know whether they can communicate with your instance or not. We don't yet have a visible way of knowing, and therefore I think it can cause confusion and frustration as I know it has for me. It has also come to my attention that knowing this helps for debugging purposes for admins, so we know whether it is a block or a bug messing things up. It saves time to know, and we can move onto making things better on our instances.<br /> <br /> I am still thinking how best to make things transparent, my intent is not to create a 'target list'... but to let my users know why they are having trouble contacting your instance if they are. That way they can decide what to do with that information, perhaps if they want to communicate then they can join another instance as well. I like to encourage people to join more than one, so they remain connected to the fediverse if something happens to their main one.<br /> <br /> I am considering creating a public list of instances that have us outright blocked and the reason for why. I also have two other lists, one for a reference point for the admins and a newer one for my personal reference point of everyone I know who can still communicate with us. I can use these two lists for clarity when people ask as well, as I know we've had a few users asking already.<br /> <br /> That is where you come in. In the interests of this list remaining civil, I want to include your reason for domain blocking (This is a full block of our domain) or domain sandboxing (This is where we don't appear on the public TL, but can communicate if your users follow us) us so I can get it accurate and neutral. I also want to hear from you if you don't, so I can add you to the list for my personal reference point which may also help in debugging issues.<br /> <br /> If you have any concerns, please get in touch with me about them.<br /> <br /> Once again, I would appreciate if people can re-post this out, or bring it to the attention of your admins. If you can point me in the right direction to communicate, that will also be helpful and make things quicker for me, especially in regards to mastodon instances. Thank you for your time and for reading this.", object2.GetSummary(), "Activity.GetObject.GetSummary")
	assert.Equal(t, "", object2.GetRepresentativeImage(), "Activity.GetObject.GetRepresentativeImage")
	assert.Equal(t, "https://social.heldscal.la/notice/1661724", object2.GetPermalink(), "Activity.GetObject.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/activity", object2.GetObjectType(), "Activity.GetObject.GetObjectType")

	object3 := object2.GetObject().(NoteLike)
	assert.NotNil(t, object3, "Activity.GetObject")
	assert.Equal(t, "tag:social.heldscal.la,2017-04-13:noticeId=1661724:objectType=note", object3.GetID(), "NoteLike.GetID")
	assert.Equal(t, "New note by sim", object3.GetName(), "NoteLike.GetName")
	assert.Equal(t, "", object3.GetSummary(), "NoteLike.GetSummary")
	assert.Equal(t, "", object3.GetRepresentativeImage(), "NoteLike.GetRepresentativeImage")
	assert.Equal(t, "https://social.heldscal.la/notice/1661724", object3.GetPermalink(), "NoteLike.GetPermalink")
	assert.Equal(t, "http://activitystrea.ms/schema/1.0/note", object3.GetObjectType(), "NoteLike.GetObjectType")
	assert.Equal(t, "Dear instance admins, I am the moderator for both sealion.club and freezepeach.xyz. I would like to reach out to you in the interests of compiling a few lists relating to whether our instances can communicate or not.<br /> <br /> It is in the interests of transparency, so that I can help my users by letting them know whether they can communicate with your instance or not. We don't yet have a visible way of knowing, and therefore I think it can cause confusion and frustration as I know it has for me. It has also come to my attention that knowing this helps for debugging purposes for admins, so we know whether it is a block or a bug messing things up. It saves time to know, and we can move onto making things better on our instances.<br /> <br /> I am still thinking how best to make things transparent, my intent is not to create a 'target list'... but to let my users know why they are having trouble contacting your instance if they are. That way they can decide what to do with that information, perhaps if they want to communicate then they can join another instance as well. I like to encourage people to join more than one, so they remain connected to the fediverse if something happens to their main one.<br /> <br /> I am considering creating a public list of instances that have us outright blocked and the reason for why. I also have two other lists, one for a reference point for the admins and a newer one for my personal reference point of everyone I know who can still communicate with us. I can use these two lists for clarity when people ask as well, as I know we've had a few users asking already.<br /> <br /> That is where you come in. In the interests of this list remaining civil, I want to include your reason for domain blocking (This is a full block of our domain) or domain sandboxing (This is where we don't appear on the public TL, but can communicate if your users follow us) us so I can get it accurate and neutral. I also want to hear from you if you don't, so I can add you to the list for my personal reference point which may also help in debugging issues.<br /> <br /> If you have any concerns, please get in touch with me about them.<br /> <br /> Once again, I would appreciate if people can re-post this out, or bring it to the attention of your admins. If you can point me in the right direction to communicate, that will also be helpful and make things quicker for me, especially in regards to mastodon instances. Thank you for your time and for reading this.", object3.GetContent(), "NoteLike.GetContent")
}
