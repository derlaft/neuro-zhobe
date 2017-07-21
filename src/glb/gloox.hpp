#include "_cgo_export.h"

#include "gloox/client.h"
#include "gloox/connectionlistener.h"
#include "gloox/mucroomhandler.h"
#include "gloox/mucroom.h"
#include "gloox/disco.h"
#include "gloox/presence.h"
#include "gloox/message.h"
#include "gloox/dataform.h"
#include "gloox/gloox.h"
#include "gloox/lastactivity.h"
#include "gloox/loghandler.h"
#include "gloox/logsink.h"
#include "gloox/stanza.h"
#include "gloox/error.h"
#include "gloox/eventhandler.h"
#include "gloox/messagesession.h"

using namespace gloox;
using namespace std;

#include <stdio.h>
#include <locale.h>
#include <string>

#include <cstdio> // [s]print[f]
#include <iostream>

class Bot : public ConnectionListener, MUCRoomHandler, LogHandler, EventHandler {
  public:

    Bot() {}
    virtual ~Bot() {}

    void start(char *uname, char *pwd, char *muc) {
      jid = new JID(uname);

      j = new Client(*jid, pwd);
      j->registerConnectionListener(this);
      j->setPresence( Presence::Available, -1 );
      j->setCompression( false );

    //  j->logInstance().registerLogHandler( LogLevelDebug, LogAreaAll, this );

      muc_jid = new JID(muc);

      m_room = new MUCRoom(j, *muc_jid, this, 0);

      if(j->connect(false)) {
        ConnectionError ce = ConnNoError;
        while(ce == ConnNoError) {
          goSched(this);
          ce = j->recv(100);
        }
      }

      // cleanup
      delete jid;
      delete muc_jid;
      delete m_room;
      delete j;
    }

    void stop() {
        if (m_room) {
            m_room->leave();
        }

        j->disconnect();

        goOnDisconnect(this, -1, 0);
    }

    char* nick() {
        if (m_room) {
            return (char*) m_room->nick().c_str();
        } 

        return (char*) "";
    }

    void reply(char *what) {
        if (!m_room) {
            return;
        }

        std::string msg(what);
        m_room->send(msg);
    }

    void reply_private(char *what, char *whom) {
        if (!m_room) {
            return;
        }

        std::string msg(what);
        JID recipient(muc_jid->bare() + "/" + whom);

        printf("wat %s\n", recipient.full().c_str());

        auto ms = new MessageSession(j, recipient);
        ms->send(msg);

        delete ms;
    }

    void ping() {
        j->xmppPing(j->jid(), this);
    }

    void kick(char *who, char *reason) {
        if (m_room) {
            std::string nickname(who);
            std::string kickReason(reason);

            m_room->kick(nickname, kickReason);
        }
    }

    virtual void handleLog( LogLevel level, LogArea area, const std::string& message ) {
      printf("log: level: %d, area: %d, %s\n", level, area, message.c_str() );
    }

    virtual void handleEvent(const Event& event) {  
        int sEvent = 1;  
        if (event.eventType() == Event::PingError) {  
                sEvent = 0;  
        }  

        goOnPing(this, sEvent);
        return;  
    }  


    virtual void onConnect() {
        if (m_room) {
            m_room->join();
        }

        goOnConnect(this);
    }

    virtual void onDisconnect(ConnectionError e) {
        goOnDisconnect(this, e, j->authError());
    }

    virtual bool onTLSConnect(const CertInfo& info) {
      return goOnTLSConnect(this, info.status);
    }

    virtual void handleMUCParticipantPresence( MUCRoom *room, const MUCRoomParticipant participant, const Presence& presence ){
        // implement automatic rejoin
        if (presence.presence() == Presence::Unavailable && participant.nick->resource() == room->nick()) {
            printf("rejoinin\n");
            this->rejoin();
        }

        // every online-ish status
        int online = presence.presence() == Presence::Available || 
            presence.presence() == Presence::Chat || 
            presence.presence() == Presence::Away || 
            presence.presence() == Presence::DND || 
            presence.presence() == Presence::XA;

        int admin = (participant.affiliation == MUCRoomAffiliation::AffiliationOwner || 
            participant.affiliation == MUCRoomAffiliation::AffiliationAdmin) &&
            (participant.role == MUCRoomRole::RoleModerator);

        auto nick = participant.nick != NULL ? (char *) participant.nick->resource().c_str() : NULL;

        goOnPresence(this, nick, online, admin);
    }

    virtual void rejoin() {
        if (m_room) {
            // try to rejoin
            // for some reason one must to leave first
            m_room->leave();
            m_room->join();
        }
    }


    virtual void handleMUCMessage(MUCRoom *room, const Message& msg, bool priv) {
      // forward to go
      goOnMessage(
              this,                                  // cobj
              (char*) msg.from().resource().c_str(), // raw_from
              (char*) msg.body().c_str(),            // raw_body
              (int) (msg.when() ? 1 : 0),            // history
              (int) priv                             // private
      );
    }

    virtual void handleMUCSubject( MUCRoom * /*room*/, const std::string& nick, const std::string& subject) {
        auto cnick = nick.empty() ? "" : nick.c_str();
        auto subj = subject.c_str();

        goOnMUCSubject(this, (char*) cnick, (char*) subj);
    }

    virtual void handleMUCError(MUCRoom * /*room*/, StanzaError error) {
        // and automatic nickname change
        if (error == StanzaError::StanzaErrorConflict) {
            m_room->setNick(m_room->nick() + "_");
            this->rejoin();
            return;
        }

        // else -- report issue
        goOnError(this, error);
    }

    virtual void handleMUCInfo(MUCRoom * /*room*/, int features, const std::string& name, const DataForm* infoForm) {
    }

    virtual void handleMUCItems( MUCRoom * /*room*/, const Disco::ItemList& items ) {
    }

    virtual void handleMUCInviteDecline( MUCRoom * /*room*/, const JID& invitee, const std::string& reason ) {
    }

    virtual bool handleMUCRoomCreation( MUCRoom *room ) {
      return true;
    }

  private:
    JID *jid;
    JID *muc_jid;
    Client *j;
    MUCRoom *m_room;
};
