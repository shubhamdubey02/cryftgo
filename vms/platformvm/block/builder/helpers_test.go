// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package builder

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/chains"
	"github.com/shubhamdubey02/cryftgo/chains/atomic"
	"github.com/shubhamdubey02/cryftgo/codec"
	"github.com/shubhamdubey02/cryftgo/codec/linearcodec"
	"github.com/shubhamdubey02/cryftgo/database"
	"github.com/shubhamdubey02/cryftgo/database/memdb"
	"github.com/shubhamdubey02/cryftgo/database/prefixdb"
	"github.com/shubhamdubey02/cryftgo/database/versiondb"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgo/snow/snowtest"
	"github.com/shubhamdubey02/cryftgo/snow/uptime"
	"github.com/shubhamdubey02/cryftgo/snow/validators"
	"github.com/shubhamdubey02/cryftgo/utils"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgo/utils/formatting"
	"github.com/shubhamdubey02/cryftgo/utils/formatting/address"
	"github.com/shubhamdubey02/cryftgo/utils/json"
	"github.com/shubhamdubey02/cryftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgo/utils/timer/mockable"
	"github.com/shubhamdubey02/cryftgo/utils/units"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/api"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/config"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/fx"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/metrics"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/network"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/reward"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/state"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/status"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs/fee"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs/mempool"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs/txstest"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/upgrade"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/utxo"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"

	blockexecutor "github.com/shubhamdubey02/cryftgo/vms/platformvm/block/executor"
	txexecutor "github.com/shubhamdubey02/cryftgo/vms/platformvm/txs/executor"
	pvalidators "github.com/shubhamdubey02/cryftgo/vms/platformvm/validators"
	walletcommon "github.com/shubhamdubey02/cryftgo/wallet/subnet/primary/common"
)

const (
	defaultWeight = 10000
	trackChecksum = false

	apricotPhase3 fork = iota
	apricotPhase5
	banff
	cortina
	durango
	eUpgrade

	latestFork = durango
)

var (
	defaultMinStakingDuration = 24 * time.Hour
	defaultMaxStakingDuration = 365 * 24 * time.Hour
	defaultGenesisTime        = time.Date(1997, 1, 1, 0, 0, 0, 0, time.UTC)
	defaultValidateStartTime  = defaultGenesisTime
	defaultValidateEndTime    = defaultValidateStartTime.Add(10 * defaultMinStakingDuration)
	defaultMinValidatorStake  = 5 * units.MilliCryft
	defaultBalance            = 100 * defaultMinValidatorStake
	preFundedKeys             = secp256k1.TestKeys()
	defaultTxFee              = uint64(100)

	testSubnet1            *txs.Tx
	testSubnet1ControlKeys = preFundedKeys[0:3]

	// Node IDs of genesis validators. Initialized in init function
	genesisNodeIDs []ids.NodeID
)

func init() {
	genesisNodeIDs = make([]ids.NodeID, len(preFundedKeys))
	for i := range preFundedKeys {
		genesisNodeIDs[i] = ids.GenerateTestNodeID()
	}
}

type fork uint8

type mutableSharedMemory struct {
	atomic.SharedMemory
}

type environment struct {
	Builder
	blkManager blockexecutor.Manager
	mempool    mempool.Mempool
	network    *network.Network
	sender     *common.SenderTest

	isBootstrapped *utils.Atomic[bool]
	config         *config.Config
	clk            *mockable.Clock
	baseDB         *versiondb.Database
	ctx            *snow.Context
	msm            *mutableSharedMemory
	fx             fx.Fx
	state          state.State
	uptimes        uptime.Manager
	utxosVerifier  utxo.Verifier
	txBuilder      *txstest.Builder
	backend        txexecutor.Backend
}

func newEnvironment(t *testing.T, f fork) *environment { //nolint:unparam
	require := require.New(t)

	res := &environment{
		isBootstrapped: &utils.Atomic[bool]{},
		config:         defaultConfig(t, f),
		clk:            defaultClock(),
	}
	res.isBootstrapped.Set(true)

	res.baseDB = versiondb.New(memdb.New())
	atomicDB := prefixdb.New([]byte{1}, res.baseDB)
	m := atomic.NewMemory(atomicDB)

	res.ctx = snowtest.Context(t, snowtest.PChainID)
	res.msm = &mutableSharedMemory{
		SharedMemory: m.NewSharedMemory(res.ctx.ChainID),
	}
	res.ctx.SharedMemory = res.msm

	res.ctx.Lock.Lock()
	defer res.ctx.Lock.Unlock()

	res.fx = defaultFx(t, res.clk, res.ctx.Log, res.isBootstrapped.Get())

	rewardsCalc := reward.NewCalculator(res.config.RewardConfig)
	res.state = defaultState(t, res.config, res.ctx, res.baseDB, rewardsCalc)

	res.uptimes = uptime.NewManager(res.state, res.clk)
	res.utxosVerifier = utxo.NewVerifier(res.ctx, res.clk, res.fx)

	res.txBuilder = txstest.NewBuilder(
		res.ctx,
		res.config,
		res.state,
	)

	genesisID := res.state.GetLastAccepted()
	res.backend = txexecutor.Backend{
		Config:       res.config,
		Ctx:          res.ctx,
		Clk:          res.clk,
		Bootstrapped: res.isBootstrapped,
		Fx:           res.fx,
		FlowChecker:  res.utxosVerifier,
		Uptimes:      res.uptimes,
		Rewards:      rewardsCalc,
	}

	registerer := prometheus.NewRegistry()
	res.sender = &common.SenderTest{T: t}
	res.sender.SendAppGossipF = func(context.Context, common.SendConfig, []byte) error {
		return nil
	}

	metrics, err := metrics.New("", registerer)
	require.NoError(err)

	res.mempool, err = mempool.New("mempool", registerer, nil)
	require.NoError(err)

	res.blkManager = blockexecutor.NewManager(
		res.mempool,
		metrics,
		res.state,
		&res.backend,
		pvalidators.TestManager,
	)

	txVerifier := network.NewLockedTxVerifier(&res.ctx.Lock, res.blkManager)
	res.network, err = network.New(
		res.backend.Ctx.Log,
		res.backend.Ctx.NodeID,
		res.backend.Ctx.SubnetID,
		res.backend.Ctx.ValidatorState,
		txVerifier,
		res.mempool,
		res.backend.Config.PartialSyncPrimaryNetwork,
		res.sender,
		registerer,
		network.DefaultConfig,
	)
	require.NoError(err)

	res.Builder = New(
		res.mempool,
		&res.backend,
		res.blkManager,
	)
	res.Builder.StartBlockTimer()

	res.blkManager.SetPreference(genesisID)
	addSubnet(t, res)

	t.Cleanup(func() {
		res.ctx.Lock.Lock()
		defer res.ctx.Lock.Unlock()

		res.Builder.ShutdownBlockTimer()

		if res.isBootstrapped.Get() {
			validatorIDs := res.config.Validators.GetValidatorIDs(constants.PrimaryNetworkID)

			require.NoError(res.uptimes.StopTracking(validatorIDs, constants.PrimaryNetworkID))

			require.NoError(res.state.Commit())
		}

		require.NoError(res.state.Close())
		require.NoError(res.baseDB.Close())
	})

	return res
}

func addSubnet(t *testing.T, env *environment) {
	require := require.New(t)

	// Create a subnet
	var err error
	testSubnet1, err = env.txBuilder.NewCreateSubnetTx(
		&secp256k1fx.OutputOwners{
			Threshold: 2,
			Addrs: []ids.ShortID{
				preFundedKeys[0].PublicKey().Address(),
				preFundedKeys[1].PublicKey().Address(),
				preFundedKeys[2].PublicKey().Address(),
			},
		},
		[]*secp256k1.PrivateKey{preFundedKeys[0]},
		walletcommon.WithChangeOwner(&secp256k1fx.OutputOwners{
			Threshold: 1,
			Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
		}),
	)
	require.NoError(err)

	// store it
	genesisID := env.state.GetLastAccepted()
	stateDiff, err := state.NewDiff(genesisID, env.blkManager)
	require.NoError(err)

	executor := txexecutor.StandardTxExecutor{
		Backend: &env.backend,
		State:   stateDiff,
		Tx:      testSubnet1,
	}
	require.NoError(testSubnet1.Unsigned.Visit(&executor))

	stateDiff.AddTx(testSubnet1, status.Committed)
	require.NoError(stateDiff.Apply(env.state))
}

func defaultState(
	t *testing.T,
	cfg *config.Config,
	ctx *snow.Context,
	db database.Database,
	rewards reward.Calculator,
) state.State {
	require := require.New(t)

	execCfg, _ := config.GetExecutionConfig([]byte(`{}`))
	genesisBytes := buildGenesisTest(t, ctx)
	state, err := state.New(
		db,
		genesisBytes,
		prometheus.NewRegistry(),
		cfg,
		execCfg,
		ctx,
		metrics.Noop,
		rewards,
	)
	require.NoError(err)

	// persist and reload to init a bunch of in-memory stuff
	state.SetHeight(0)
	require.NoError(state.Commit())
	return state
}

func defaultConfig(t *testing.T, f fork) *config.Config {
	c := &config.Config{
		Chains:                 chains.TestManager,
		UptimeLockedCalculator: uptime.NewLockedCalculator(),
		Validators:             validators.NewManager(),
		StaticFeeConfig: fee.StaticConfig{
			TxFee:                 defaultTxFee,
			CreateSubnetTxFee:     100 * defaultTxFee,
			CreateBlockchainTxFee: 100 * defaultTxFee,
		},
		MinValidatorStake: 5 * units.MilliCryft,
		MaxValidatorStake: 500 * units.MilliCryft,
		MinDelegatorStake: 1 * units.MilliCryft,
		MinStakeDuration:  defaultMinStakingDuration,
		MaxStakeDuration:  defaultMaxStakingDuration,
		RewardConfig: reward.Config{
			MaxConsumptionRate: .12 * reward.PercentDenominator,
			MinConsumptionRate: .10 * reward.PercentDenominator,
			MintingPeriod:      365 * 24 * time.Hour,
			SupplyCap:          720 * units.MegaCryft,
		},
		UpgradeConfig: upgrade.Config{
			ApricotPhase3Time: mockable.MaxTime,
			ApricotPhase5Time: mockable.MaxTime,
			BanffTime:         mockable.MaxTime,
			CortinaTime:       mockable.MaxTime,
			DurangoTime:       mockable.MaxTime,
			EUpgradeTime:      mockable.MaxTime,
		},
	}

	switch f {
	case eUpgrade:
		c.UpgradeConfig.EUpgradeTime = time.Time{} // neglecting fork ordering this for package tests
		fallthrough
	case durango:
		c.UpgradeConfig.DurangoTime = time.Time{} // neglecting fork ordering for this package's tests
		fallthrough
	case cortina:
		c.UpgradeConfig.CortinaTime = time.Time{} // neglecting fork ordering for this package's tests
		fallthrough
	case banff:
		c.UpgradeConfig.BanffTime = time.Time{} // neglecting fork ordering for this package's tests
		fallthrough
	case apricotPhase5:
		c.UpgradeConfig.ApricotPhase5Time = defaultValidateEndTime
		fallthrough
	case apricotPhase3:
		c.UpgradeConfig.ApricotPhase3Time = defaultValidateEndTime
	default:
		require.FailNow(t, "unhandled fork", f)
	}

	return c
}

func defaultClock() *mockable.Clock {
	// set time after Banff fork (and before default nextStakerTime)
	clk := &mockable.Clock{}
	clk.Set(defaultGenesisTime)
	return clk
}

type fxVMInt struct {
	registry codec.Registry
	clk      *mockable.Clock
	log      logging.Logger
}

func (fvi *fxVMInt) CodecRegistry() codec.Registry {
	return fvi.registry
}

func (fvi *fxVMInt) Clock() *mockable.Clock {
	return fvi.clk
}

func (fvi *fxVMInt) Logger() logging.Logger {
	return fvi.log
}

func defaultFx(t *testing.T, clk *mockable.Clock, log logging.Logger, isBootstrapped bool) fx.Fx {
	require := require.New(t)

	fxVMInt := &fxVMInt{
		registry: linearcodec.NewDefault(),
		clk:      clk,
		log:      log,
	}
	res := &secp256k1fx.Fx{}
	require.NoError(res.Initialize(fxVMInt))
	if isBootstrapped {
		require.NoError(res.Bootstrapped())
	}
	return res
}

func buildGenesisTest(t *testing.T, ctx *snow.Context) []byte {
	require := require.New(t)

	genesisUTXOs := make([]api.UTXO, len(preFundedKeys))
	for i, key := range preFundedKeys {
		id := key.PublicKey().Address()
		addr, err := address.FormatBech32(constants.UnitTestHRP, id.Bytes())
		require.NoError(err)
		genesisUTXOs[i] = api.UTXO{
			Amount:  json.Uint64(defaultBalance),
			Address: addr,
		}
	}

	genesisValidators := make([]api.GenesisPermissionlessValidator, len(genesisNodeIDs))
	for i, nodeID := range genesisNodeIDs {
		addr, err := address.FormatBech32(constants.UnitTestHRP, nodeID.Bytes())
		require.NoError(err)
		genesisValidators[i] = api.GenesisPermissionlessValidator{
			GenesisValidator: api.GenesisValidator{
				StartTime: json.Uint64(defaultValidateStartTime.Unix()),
				EndTime:   json.Uint64(defaultValidateEndTime.Unix()),
				NodeID:    nodeID,
			},
			RewardOwner: &api.Owner{
				Threshold: 1,
				Addresses: []string{addr},
			},
			Staked: []api.UTXO{{
				Amount:  json.Uint64(defaultWeight),
				Address: addr,
			}},
			DelegationFee: reward.PercentDenominator,
		}
	}

	buildGenesisArgs := api.BuildGenesisArgs{
		NetworkID:     json.Uint32(constants.UnitTestID),
		CryftAssetID:  ctx.CRYFTAssetID,
		UTXOs:         genesisUTXOs,
		Validators:    genesisValidators,
		Chains:        nil,
		Time:          json.Uint64(defaultGenesisTime.Unix()),
		InitialSupply: json.Uint64(360 * units.MegaCryft),
		Encoding:      formatting.Hex,
	}

	buildGenesisResponse := api.BuildGenesisReply{}
	platformvmSS := api.StaticService{}
	require.NoError(platformvmSS.BuildGenesis(nil, &buildGenesisArgs, &buildGenesisResponse))

	genesisBytes, err := formatting.Decode(buildGenesisResponse.Encoding, buildGenesisResponse.Bytes)
	require.NoError(err)

	return genesisBytes
}
