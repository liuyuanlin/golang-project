#ifndef _FightCalculator_H_
#define _FightCalculator_H_
#include "msg.pb.h"
#include "LogicBuilding.h"
#include "LogicClanComponent.h"

//debug战斗不匹配
#ifndef _SUPER_DEBUG_
#define _SUPER_DEBUG_
#endif

struct BattleResult
{
    uint64_t            _playerlid;
    unsigned int        _goldStolen;
    unsigned int        _foodStolen;
    float               _damagePercent;
    unsigned int        _stars;
    int                 _trophy;
    unsigned int        _exp;
    unsigned int        _wuhun;
    rpc::VillageInfo    m_vi;
};
struct AllianceArmy
{
    unsigned int        _id;
    unsigned char       _level;
    rpc::CharacterType  _type;
};
class FightCalculator : public GameLogic::BuildingListener, public GameLogic::BattleInfoChangeListener,public GameLogic::AllianceArmyDeployListener
{
public:
    static FightCalculator* Instance();
    BattleResult CalculateBattleResult(rpc::AttackBegin& ab);
    BattleResult CalculateBattleResult(rpc::PVEAttackBegin& ab);
    
    virtual void OnAddRobGoldCount(unsigned int addRobGoldCount);
    virtual void OnAddRobFoodCount(unsigned int addRobFoodCount);
    virtual void OnStarCountChange(unsigned int starCount);
    virtual void On100PercentDestroy(unsigned char addStarCount);
    virtual void OnArmyAllDead();
    virtual void OnCenterDead(unsigned int centerLevel);
    
    virtual void BombDead(string cateName,unsigned int index);
    void ReadMissionConfigList(string missinConfig);
    void OnAllianceArmyDeploy(GameLogic::LogicAvatar* avatar,rpc::CharacterType type,int level,GameLogic::Float2 isoPos);
private:
    FightCalculator();
    BattleResult CalculateBattleResult(rpc::VillageInfo* vInfo,
                                       const ::google::protobuf::RepeatedPtrField< ::rpc::AttackerInfo>& attsInfo,
                                       const ::google::protobuf::RepeatedPtrField< ::rpc::SpellInfo>& spellsInfo,
                                       const ::rpc::ClanForceInfo* allianceArmy,
                                       int src_trophy,int tar_trophy,
                                       int totalTickCount,uint64_t playerId,bool bRestoreVillage,
                                       bool bUseExternalRes = false,int goldCount = 0,int foodCount = 0,
                                       int diamond = 0);
    unsigned int m_CenterLevel;
    rpc::VillageInfo*       m_pPVPVillageInfo;
    map<string,rpc::VillageInfo*>       m_PVEVillageMap;
    map<unsigned int,vector<AllianceArmy> >      m_AllianceCharacterMap;
};

#endif
