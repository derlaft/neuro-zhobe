//cfoo.cpp
#include "gloox.hpp"
#include "gloox.h"

GBot BotInit() {
    Bot* ret = new Bot();
    return (void *)ret;
}

void BotConnect(GBot b, char *jid, char *pwd, char *room) {
    auto bot = (Bot*) b;
    bot->start(jid, pwd, room);
    delete jid, pwd, room;
}

void BotDisconnect(GBot b) {
    auto bot = (Bot*) b;
    bot->stop();
}

void BotFree(GBot b) {
    auto bot = (Bot *) b;
    delete bot;
}

void BotReply(GBot b, char *what) {
    auto bot = (Bot *) b;
    bot->reply(what);
    delete what;
}

void BotReplyPrivate(GBot b, char *what, char *whom) {
    auto bot = (Bot *) b;
    bot->reply_private(what, whom);
    delete what, whom;
}

void BotKick(GBot b, char *who, char *reason) {
    auto bot = (Bot *) b;
    bot->kick(who, reason);
    delete who, reason;
}

char *BotNick(GBot b) {
    auto bot = (Bot *) b;
    return bot->nick();
}

void BotPingRoom(GBot b) {
    auto bot = (Bot *) b;
    bot->ping();
}
