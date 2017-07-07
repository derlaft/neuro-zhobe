#ifdef __cplusplus
extern "C" {
#endif

    // Bot
    typedef void* GBot;
    GBot BotInit(void);
    void BotFree(GBot);
    void BotConnect(GBot, char*, char*, char*);
    void BotDisconnect(GBot);
    void BotReply(GBot, char*);
    void BotKick(GBot, char*, char*);
    char* BotNick(GBot);
    void BotPingRoom(GBot);

#ifdef __cplusplus
};
#endif
