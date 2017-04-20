package activitystreams

const fixtureImpliedActivity1 = `
<?xml version="1.0" encoding="UTF-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom" xmlns:activity="http://activitystrea.ms/spec/1.0/" xmlns:georss="http://www.georss.org/georss" xmlns:media="http://purl.org/syndication/atommedia" xmlns:ostatus="http://ostatus.org/schema/1.0" xmlns:poco="http://portablecontacts.net/spec/1.0" xmlns:statusnet="http://status.net/schema/api/1/" xmlns:thr="http://purl.org/syndication/thread/1.0">
	<generator uri="https://gnu.io/social" version="1.2.0-beta4">GNU social</generator>
	<id>https://quitter.im/api/statuses/user_timeline/235.atom</id>
	<title>newsstream timeline</title>
	<subtitle>Updates from newsstream on Quitter IM!</subtitle>
	<logo>https://quitter.im/avatar/235-96-20151207175006.jpeg</logo>
	<updated>2017-04-15T04:12:30+00:00</updated>
	<author>
		<activity:object-type>http://activitystrea.ms/schema/1.0/person</activity:object-type>
		<uri>https://quitter.im/user/235</uri>
		<name>newsstream</name>
		<summary>Realtime News</summary>
		<link href="https://quitter.im/newsstream" rel="alternate" type="text/html"/>
		<link href="https://quitter.im/avatar/235-original-20151207175006.jpeg" media:height="1040" media:width="1040" rel="avatar" type="image/jpeg"/>
		<link href="https://quitter.im/avatar/235-96-20151207175006.jpeg" media:height="96" media:width="96" rel="avatar" type="image/jpeg"/>
		<link href="https://quitter.im/avatar/235-48-20151207175006.jpeg" media:height="48" media:width="48" rel="avatar" type="image/jpeg"/>
		<link href="https://quitter.im/avatar/235-24-20151207175006.jpeg" media:height="24" media:width="24" rel="avatar" type="image/jpeg"/>
		<poco:preferredUsername>newsstream</poco:preferredUsername>
		<poco:displayName>NewsStream</poco:displayName>
		<poco:note>Realtime News</poco:note>
		<followers url="https://quitter.im/newsstream/subscribers"/>
		<statusnet:profile_info local_id="235"/>
	</author>
	<link href="https://quitter.im/newsstream" rel="alternate" type="text/html"/>
	<link href="https://quitter.im/main/sup" rel="http://api.friendfeed.com/2008/03#sup" type="application/json"/>
	<link href="https://quitter.im/main/push/hub" rel="hub"/>
	<link href="https://quitter.im/main/salmon/user/235" rel="salmon"/>
	<link href="https://quitter.im/main/salmon/user/235" rel="http://salmon-protocol.org/ns/salmon-replies"/>
	<link href="https://quitter.im/main/salmon/user/235" rel="http://salmon-protocol.org/ns/salmon-mention"/>
	<link href="https://quitter.im/api/statuses/user_timeline/235.atom" rel="self" type="application/atom+xml"/>
	<entry>
		<activity:object-type>http://activitystrea.ms/schema/1.0/note</activity:object-type>
		<id>tag:quitter.im,2017-04-15:noticeId=1415216:objectType=note</id>
		<title>New note by newsstream</title>
		<content type="html">North Korean leader Kim Jong-Un watches ranks of goose-stepping soldiers, tanks and hardware parade in Pyongyang…  &lt;a href=&quot;https://twitter.com/i/web/status/853098655493967872&quot; title=&quot;https://twitter.com/i/web/status/853098655493967872&quot; class=&quot;attachment&quot; id=&quot;attachment-460600&quot; rel=&quot;nofollow external&quot;&gt;https://twitter.com/i/web/status/853098655493967872&lt;/a&gt;</content>
		<link href="https://quitter.im/notice/1415216" rel="alternate" type="text/html"/>
		<status_net notice_id="1415216"/>
		<activity:verb>http://activitystrea.ms/schema/1.0/post</activity:verb>
		<published>2017-04-15T04:12:24+00:00</published>
		<updated>2017-04-15T04:12:24+00:00</updated>
		<link href="https://quitter.im/conversation/778833" rel="ostatus:conversation"/>
		<ostatus:conversation href="https://quitter.im/conversation/778833" local_id="778833" ref="tag:quitter.im,2017-04-15:objectType=thread:nonce=d2d46dc8b3f9d7ac">tag:quitter.im,2017-04-15:objectType=thread:nonce=d2d46dc8b3f9d7ac</ostatus:conversation>
		<link href="http://activityschema.org/collection/public" ostatus:object-type="http://activitystrea.ms/schema/1.0/collection" rel="mentioned"/>
		<link href="https://quitter.im/api/statuses/show/1415216.atom" rel="self" type="application/atom+xml"/>
		<link href="https://quitter.im/api/statuses/show/1415216.atom" rel="edit" type="application/atom+xml"/>
		<statusnet:notice_info local_id="1415216" source="api"/>
	</entry>
</feed>
`

const fixtureImpliedActivity2 = `
<?xml version="1.0"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:activity="http://activitystrea.ms/spec/1.0/" xmlns:mastodon="http://mastodon.social/schema/1.0" xmlns:media="http://purl.org/syndication/atommedia" xmlns:ostatus="http://ostatus.org/schema/1.0" xmlns:poco="http://portablecontacts.net/spec/1.0" xmlns:thr="http://purl.org/syndication/thread/1.0">
  <id>https://social.tchncs.de/users/milan.atom</id>
  <title>✨ Milan 🐘</title>
  <subtitle>Odd humor, in love with administrating Linux servers.
Wanna help me out a bit? https://tchncs.de/donate</subtitle>
  <updated>2017-04-10T21:24:16Z</updated>
  <logo>https://social.tchncs.de/system/accounts/avatars/000/001/833/original/6a2fe0e775c2baf2.jpg?1491589004</logo>
  <author>
    <id>https://social.tchncs.de/users/milan</id>
    <activity:object-type>http://activitystrea.ms/schema/1.0/person</activity:object-type>
    <uri>https://social.tchncs.de/users/milan</uri>
    <name>milan</name>
    <email>milan@social.tchncs.de</email>
    <summary>Odd humor, in love with administrating Linux servers.
Wanna help me out a bit? https://tchncs.de/donate</summary>
    <link href="https://social.tchncs.de/@milan" rel="alternate" type="text/html"/>
    <link href="https://social.tchncs.de/system/accounts/avatars/000/001/833/original/6a2fe0e775c2baf2.jpg?1491589004" media:height="120" media:width="120" rel="avatar" type="image/jpeg"/>
    <link href="https://social.tchncs.de/system/accounts/headers/000/001/833/original/media.jpg?1485591778" media:height="335" media:width="700" rel="header" type="image/jpeg"/>
    <poco:preferredUsername>milan</poco:preferredUsername>
    <poco:displayName>✨ Milan 🐘</poco:displayName>
    <poco:note>Odd humor, in love with administrating Linux servers.
Wanna help me out a bit? &lt;a href=&quot;https://tchncs.de/donate&quot; rel=&quot;nofollow noopener&quot; target=&quot;_blank&quot;&gt;&lt;span class=&quot;invisible&quot;&gt;https://&lt;/span&gt;&lt;span class=&quot;&quot;&gt;tchncs.de/donate&lt;/span&gt;&lt;span class=&quot;invisible&quot;&gt;&lt;/span&gt;&lt;/a&gt;</poco:note>
    <mastodon:scope>public</mastodon:scope>
  </author>
  <link href="https://social.tchncs.de/@milan" rel="alternate" type="text/html"/>
  <link href="https://social.tchncs.de/users/milan.atom" rel="self" type="application/atom+xml"/>
  <link href="https://social.tchncs.de/api/push" rel="hub"/>
  <link href="https://social.tchncs.de/api/salmon/1833" rel="salmon"/>
  <entry>
    <id>tag:social.tchncs.de,2017-04-15:objectId=755238:objectType=Status</id>
    <published>2017-04-15T06:59:43Z</published>
    <updated>2017-04-15T06:59:43Z</updated>
    <title>New status by milan</title>
    <activity:object-type>http://activitystrea.ms/schema/1.0/note</activity:object-type>
    <activity:verb>http://activitystrea.ms/schema/1.0/post</activity:verb>
    <content type="html">&lt;p&gt;Ui i should have had updated my spf record after adding ipv6 to this server: &lt;a href=&quot;http://www.openspf.org/Why?s=helo;id=mail.tchncs.de;ip=2a00%3Af820%3A417%3A%3A5e51%3Ac752;r=mail.privater-postbote.de&quot; rel=&quot;nofollow noopener&quot; target=&quot;_blank&quot;&gt;&lt;span class=&quot;invisible&quot;&gt;http://www.&lt;/span&gt;&lt;span class=&quot;ellipsis&quot;&gt;openspf.org/Why?s=helo;id=mail&lt;/span&gt;&lt;span class=&quot;invisible&quot;&gt;.tchncs.de;ip=2a00%3Af820%3A417%3A%3A5e51%3Ac752;r=mail.privater-postbote.de&lt;/span&gt;&lt;/a&gt;&lt;/p&gt;&lt;p&gt;Hey other mastodon servers i hope you have a proper mailsetup in the first place!&lt;/p&gt;</content>
    <link href="http://activityschema.org/collection/public" ostatus:object-type="http://activitystrea.ms/schema/1.0/collection" rel="mentioned"/>
    <mastodon:scope>public</mastodon:scope>
    <link href="https://social.tchncs.de/users/milan/updates/45844" rel="alternate" type="text/html"/>
    <link href="https://social.tchncs.de/users/milan/updates/45844.atom" rel="self" type="application/atom+xml"/>
  </entry>
</feed>
`

const fixtureNested = `
<?xml version="1.0" encoding="UTF-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom" xmlns:activity="http://activitystrea.ms/spec/1.0/" xmlns:georss="http://www.georss.org/georss" xmlns:media="http://purl.org/syndication/atommedia" xmlns:ostatus="http://ostatus.org/schema/1.0" xmlns:poco="http://portablecontacts.net/spec/1.0" xmlns:statusnet="http://status.net/schema/api/1/" xmlns:thr="http://purl.org/syndication/thread/1.0">
  <generator uri="https://gnu.io/social" version="1.0.0-dev">GNU social</generator>
  <id>https://social.heldscal.la/api/statuses/user_timeline/23211.atom</id>
  <title>lambadalambda timeline</title>
  <subtitle>Updates from lambadalambda on social.heldscal.la!</subtitle>
  <logo>https://social.heldscal.la/avatar/23211-96-20170205203241.jpeg</logo>
  <updated>2017-04-13T06:57:43+00:00</updated>
  <author>
    <activity:object-type>http://activitystrea.ms/schema/1.0/person</activity:object-type>
    <uri>https://social.heldscal.la/user/23211</uri>
    <name>lambadalambda</name>
    <summary>Call me Deacon Blues.</summary>
    <link href="https://social.heldscal.la/lambadalambda" rel="alternate" type="text/html"/>
    <link href="https://social.heldscal.la/avatar/23211-236-20170205203241.jpeg" media:height="236" media:width="236" rel="avatar" type="image/jpeg"/>
    <link href="https://social.heldscal.la/avatar/23211-96-20170205203241.jpeg" media:height="96" media:width="96" rel="avatar" type="image/jpeg"/>
    <link href="https://social.heldscal.la/avatar/23211-48-20170205203253.jpeg" media:height="48" media:width="48" rel="avatar" type="image/jpeg"/>
    <link href="https://social.heldscal.la/avatar/23211-24-20170205203309.jpeg" media:height="24" media:width="24" rel="avatar" type="image/jpeg"/>
    <poco:preferredUsername>lambadalambda</poco:preferredUsername>
    <poco:displayName>Constance Variable</poco:displayName>
    <poco:note>Call me Deacon Blues.</poco:note>
    <poco:address>
      <poco:formatted>Berlin</poco:formatted>
    </poco:address>
    <poco:urls>
      <poco:type>homepage</poco:type>
      <poco:value>https://heldscal.la</poco:value>
      <poco:primary>true</poco:primary>
    </poco:urls>
    <followers url="https://social.heldscal.la/lambadalambda/subscribers"/>
    <statusnet:profile_info local_id="23211"/>
  </author>
  <link href="https://social.heldscal.la/lambadalambda" rel="alternate" type="text/html"/>
  <link href="https://social.heldscal.la/main/sup" rel="http://api.friendfeed.com/2008/03#sup" type="application/json"/>
  <link href="https://social.heldscal.la/main/push/hub" rel="hub"/>
  <link href="https://social.heldscal.la/main/salmon/user/23211" rel="salmon"/>
  <link href="https://social.heldscal.la/main/salmon/user/23211" rel="http://salmon-protocol.org/ns/salmon-replies"/>
  <link href="https://social.heldscal.la/main/salmon/user/23211" rel="http://salmon-protocol.org/ns/salmon-mention"/>
  <link href="https://social.heldscal.la/api/statuses/user_timeline/23211.atom" rel="self" type="application/atom+xml"/>
  <entry>
    <id>tag:social.heldscal.la,2017-04-13:noticeId=1661855:objectType=note</id>
    <title>lambadalambda repeated a notice by lohang</title>
    <content type="html">RT @&lt;a href=&quot;https://gnusocial.no/user/695&quot; class=&quot;h-card u-url p-nickname mention&quot; title=&quot;Lohan G&quot;&gt;lohang&lt;/a&gt; RT @&lt;a href=&quot;https://sealion.club/user/579&quot; class=&quot;h-card u-url p-nickname mention&quot; title=&quot;Sim Bot&quot;&gt;sim&lt;/a&gt; Dear instance admins, I am the moderator for both sealion.club and freezepeach.xyz. I would like to reach out to you in the interests of compiling a few lists relating to whether our instances can communicate or not. It is in the interests of transparency, so that I can help my users by letting them know whether they can communicate with your instance or not. We don't yet have a visible way of knowing, and therefore I think it can cause confusion and frustration as I know it has for me. It has also come to my attention that knowing this helps for debugging purposes for admins, so we know whether it is a block or a bug messing things up. It saves time to know, and we can move onto making things better on our instances. I am still thinking how best to make things transparent, my intent is not to create a 'target list'... but to let my users know why they are having trouble contacting your instance if they are. That way they can decide what to do with that information, perhaps ...</content>
    <link href="https://social.heldscal.la/notice/1661855" rel="alternate" type="text/html"/>
    <activity:verb>http://activitystrea.ms/schema/1.0/share</activity:verb>
    <published>2017-04-13T06:57:43+00:00</published>
    <updated>2017-04-13T06:57:43+00:00</updated>
    <activity:object>
      <activity:object-type>http://activitystrea.ms/schema/1.0/activity</activity:object-type>
      <id>tag:gnusocial.no,2017-04-13:noticeId=1938446:objectType=note</id>
      <title>lohang repeated a notice by sim</title>
      <content type="html">RT @&lt;a href=&quot;https://social.heldscal.la/user/26137&quot; class=&quot;h-card mention&quot; title=&quot;Sim&quot;&gt;sim&lt;/a&gt; Dear instance admins, I am the moderator for both sealion.club and freezepeach.xyz. I would like to reach out to you in the interests of compiling a few lists relating to whether our instances can communicate or not. It is in the interests of transparency, so that I can help my users by letting them know whether they can communicate with your instance or not. We don't yet have a visible way of knowing, and therefore I think it can cause confusion and frustration as I know it has for me. It has also come to my attention that knowing this helps for debugging purposes for admins, so we know whether it is a block or a bug messing things up. It saves time to know, and we can move onto making things better on our instances. I am still thinking how best to make things transparent, my intent is not to create a 'target list'... but to let my users know why they are having trouble contacting your instance if they are. That way they can decide what to do with that information, perhaps ...</content>
      <link href="https://gnusocial.no/notice/1938446" rel="alternate" type="text/html"/>
      <activity:verb>http://activitystrea.ms/schema/1.0/share</activity:verb>
      <published>2017-04-13T06:46:16+00:00</published>
      <updated>2017-04-13T06:46:16+00:00</updated>
      <author>
        <activity:object-type>http://activitystrea.ms/schema/1.0/person</activity:object-type>
        <uri>https://gnusocial.no/user/695</uri>
        <name>lohang</name>
        <summary>Visual artist, writer and translator. Loves theatre and performance art. A free software user</summary>
        <link href="https://gnusocial.no/lohang" rel="alternate" type="text/html"/>
        <link href="https://social.heldscal.la/avatar/24350-original-20170223015616.jpeg" media:height="96" media:width="96" rel="avatar" type="image/jpeg"/>
        <link href="https://social.heldscal.la/avatar/24350-original-20170223015616.jpeg" media:height="96" media:width="96" rel="avatar" type="image/jpeg"/>
        <link href="https://social.heldscal.la/avatar/24350-48-20170223015618.jpeg" media:height="48" media:width="48" rel="avatar" type="image/jpeg"/>
        <link href="https://social.heldscal.la/avatar/24350-24-20170223023313.jpeg" media:height="24" media:width="24" rel="avatar" type="image/jpeg"/>
        <poco:preferredUsername>lohang</poco:preferredUsername>
        <poco:displayName>Lohan G</poco:displayName>
        <poco:note>Visual artist, writer and translator. Loves theatre and performance art. A free software user</poco:note>
        <poco:address>
          <poco:formatted>Sri Lanka</poco:formatted>
        </poco:address>
        <poco:urls>
          <poco:type>homepage</poco:type>
          <poco:value>https://lohangunaweera.tk</poco:value>
          <poco:primary>true</poco:primary>
        </poco:urls>
        <statusnet:profile_info local_id="24350"/>
      </author>
      <activity:object>
        <activity:object-type>http://activitystrea.ms/schema/1.0/activity</activity:object-type>
        <id>tag:social.heldscal.la,2017-04-13:noticeId=1661724:objectType=note</id>
        <title/>
        <content type="html">Dear instance admins, I am the moderator for both sealion.club and freezepeach.xyz. I would like to reach out to you in the interests of compiling a few lists relating to whether our instances can communicate or not.&lt;br /&gt; &lt;br /&gt; It is in the interests of transparency, so that I can help my users by letting them know whether they can communicate with your instance or not. We don't yet have a visible way of knowing, and therefore I think it can cause confusion and frustration as I know it has for me. It has also come to my attention that knowing this helps for debugging purposes for admins, so we know whether it is a block or a bug messing things up. It saves time to know, and we can move onto making things better on our instances.&lt;br /&gt; &lt;br /&gt; I am still thinking how best to make things transparent, my intent is not to create a 'target list'... but to let my users know why they are having trouble contacting your instance if they are. That way they can decide what to do with that information, perhaps if they want to communicate then they can join another instance as well. I like to encourage people to join more than one, so they remain connected to the fediverse if something happens to their main one.&lt;br /&gt; &lt;br /&gt; I am considering creating a public list of instances that have us outright blocked and the reason for why. I also have two other lists, one for a reference point for the admins and a newer one for my personal reference point of everyone I know who can still communicate with us. I can use these two lists for clarity when people ask as well, as I know we've had a few users asking already.&lt;br /&gt; &lt;br /&gt; That is where you come in. In the interests of this list remaining civil, I want to include your reason for domain blocking (This is a full block of our domain) or domain sandboxing (This is where we don't appear on the public TL, but can communicate if your users follow us) us so I can get it accurate and neutral. I also want to hear from you if you don't, so I can add you to the list for my personal reference point which may also help in debugging issues.&lt;br /&gt; &lt;br /&gt; If you have any concerns, please get in touch with me about them.&lt;br /&gt; &lt;br /&gt; Once again, I would appreciate if people can re-post this out, or bring it to the attention of your admins. If you can point me in the right direction to communicate, that will also be helpful and make things quicker for me, especially in regards to mastodon instances. Thank you for your time and for reading this.</content>
        <link href="https://social.heldscal.la/notice/1661724" rel="alternate" type="text/html"/>
        <activity:verb>http://activitystrea.ms/schema/1.0/post</activity:verb>
        <published>2017-04-13T06:45:59+00:00</published>
        <updated>2017-04-13T06:45:59+00:00</updated>
        <author>
          <activity:object-type>http://activitystrea.ms/schema/1.0/person</activity:object-type>
          <uri>https://social.heldscal.la/user/26137</uri>
          <name>sim</name>
          <summary>Hello wider fediverse!</summary>
          <link href="https://social.heldscal.la/sim" rel="alternate" type="text/html"/>
          <link href="https://social.heldscal.la/avatar/26137-96-20161109191319.jpeg" media:height="96" media:width="96" rel="avatar" type="image/jpeg"/>
          <link href="https://social.heldscal.la/avatar/26137-48-20161109191325.jpeg" media:height="48" media:width="48" rel="avatar" type="image/jpeg"/>
          <link href="https://social.heldscal.la/avatar/26137-24-20161109191606.jpeg" media:height="24" media:width="24" rel="avatar" type="image/jpeg"/>
          <poco:preferredUsername>sim</poco:preferredUsername>
          <poco:displayName>Sim</poco:displayName>
          <poco:note>Hello wider fediverse!</poco:note>
          <poco:address>
            <poco:formatted>SimCity</poco:formatted>
          </poco:address>
          <followers url="https://social.heldscal.la/sim/subscribers"/>
          <statusnet:profile_info local_id="26137"/>
        </author>
        <activity:object>
          <activity:object-type>http://activitystrea.ms/schema/1.0/note</activity:object-type>
          <id>tag:social.heldscal.la,2017-04-13:noticeId=1661724:objectType=note</id>
          <title>New note by sim</title>
          <content type="html">Dear instance admins, I am the moderator for both sealion.club and freezepeach.xyz. I would like to reach out to you in the interests of compiling a few lists relating to whether our instances can communicate or not.&lt;br /&gt; &lt;br /&gt; It is in the interests of transparency, so that I can help my users by letting them know whether they can communicate with your instance or not. We don't yet have a visible way of knowing, and therefore I think it can cause confusion and frustration as I know it has for me. It has also come to my attention that knowing this helps for debugging purposes for admins, so we know whether it is a block or a bug messing things up. It saves time to know, and we can move onto making things better on our instances.&lt;br /&gt; &lt;br /&gt; I am still thinking how best to make things transparent, my intent is not to create a 'target list'... but to let my users know why they are having trouble contacting your instance if they are. That way they can decide what to do with that information, perhaps if they want to communicate then they can join another instance as well. I like to encourage people to join more than one, so they remain connected to the fediverse if something happens to their main one.&lt;br /&gt; &lt;br /&gt; I am considering creating a public list of instances that have us outright blocked and the reason for why. I also have two other lists, one for a reference point for the admins and a newer one for my personal reference point of everyone I know who can still communicate with us. I can use these two lists for clarity when people ask as well, as I know we've had a few users asking already.&lt;br /&gt; &lt;br /&gt; That is where you come in. In the interests of this list remaining civil, I want to include your reason for domain blocking (This is a full block of our domain) or domain sandboxing (This is where we don't appear on the public TL, but can communicate if your users follow us) us so I can get it accurate and neutral. I also want to hear from you if you don't, so I can add you to the list for my personal reference point which may also help in debugging issues.&lt;br /&gt; &lt;br /&gt; If you have any concerns, please get in touch with me about them.&lt;br /&gt; &lt;br /&gt; Once again, I would appreciate if people can re-post this out, or bring it to the attention of your admins. If you can point me in the right direction to communicate, that will also be helpful and make things quicker for me, especially in regards to mastodon instances. Thank you for your time and for reading this.</content>
          <link href="https://social.heldscal.la/notice/1661724" rel="alternate" type="text/html"/>
          <status_net notice_id="1661724"/>
        </activity:object>
        <link href="https://social.heldscal.la/conversation/830581" rel="ostatus:conversation"/>
        <ostatus:conversation href="https://social.heldscal.la/conversation/830581" local_id="830581" ref="tag:social.heldscal.la,2017-04-13:objectType=thread:nonce=0215db01cec54295">tag:social.heldscal.la,2017-04-13:objectType=thread:nonce=0215db01cec54295</ostatus:conversation>
        <link href="http://activityschema.org/collection/public" ostatus:object-type="http://activitystrea.ms/schema/1.0/collection" rel="mentioned"/>
        <source>
          <id>https://social.heldscal.la/api/statuses/user_timeline/26137.atom</id>
          <title>Sim</title>
          <link href="https://social.heldscal.la/sim" rel="alternate" type="text/html"/>
          <link href="https://social.heldscal.la/api/statuses/user_timeline/26137.atom" rel="self" type="application/atom+xml"/>
          <link href="https://creativecommons.org/licenses/by/3.0/" rel="license"/>
          <icon>https://social.heldscal.la/avatar/26137-96-20161109191319.jpeg</icon>
          <updated>2017-04-13T06:51:59+00:00</updated>
        </source>
        <link href="https://social.heldscal.la/api/statuses/show/1661724.atom" rel="self" type="application/atom+xml"/>
        <link href="https://social.heldscal.la/api/statuses/show/1661724.atom" rel="edit" type="application/atom+xml"/>
      </activity:object>
      <link href="https://social.heldscal.la/conversation/830581" rel="ostatus:conversation"/>
      <ostatus:conversation href="https://social.heldscal.la/conversation/830581" local_id="830581" ref="tag:social.heldscal.la,2017-04-13:objectType=thread:nonce=0215db01cec54295">tag:social.heldscal.la,2017-04-13:objectType=thread:nonce=0215db01cec54295</ostatus:conversation>
      <link href="http://activityschema.org/collection/public" ostatus:object-type="http://activitystrea.ms/schema/1.0/collection" rel="mentioned"/>
      <source>
        <id>https://gnusocial.no/api/statuses/user_timeline/695.atom</id>
        <title>Lohan G</title>
        <link href="https://gnusocial.no/lohang" rel="alternate" type="text/html"/>
        <link href="https://gnusocial.no/api/statuses/user_timeline/695.atom" rel="self" type="application/atom+xml"/>
        <icon>https://social.heldscal.la/avatar/24350-original-20170223015616.jpeg</icon>
        <updated>2017-04-13T06:46:16+00:00</updated>
      </source>
    </activity:object>
    <link href="https://social.heldscal.la/conversation/830581" rel="ostatus:conversation"/>
    <ostatus:conversation href="https://social.heldscal.la/conversation/830581" local_id="830581" ref="tag:social.heldscal.la,2017-04-13:objectType=thread:nonce=0215db01cec54295">tag:social.heldscal.la,2017-04-13:objectType=thread:nonce=0215db01cec54295</ostatus:conversation>
    <link href="http://activityschema.org/collection/public" ostatus:object-type="http://activitystrea.ms/schema/1.0/collection" rel="mentioned"/>
    <link href="https://social.heldscal.la/api/statuses/show/1661855.atom" rel="self" type="application/atom+xml"/>
    <link href="https://social.heldscal.la/api/statuses/show/1661855.atom" rel="edit" type="application/atom+xml"/>
    <statusnet:notice_info local_id="1661855" repeat_of="1661791" source="api"/>
  </entry>
</feed>
`
