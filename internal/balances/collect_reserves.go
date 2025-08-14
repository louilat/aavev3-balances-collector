package balances

import (
	"aavev3-raw-balances-collector/internal/pool"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
)

type ReserveData struct {
	BlockNumber             *big.Int `json:"blockNumber"`
	Name                    string   `json:"name"`
	UnderlyingAsset         string   `json:"underlyingAsset"`
	Decimals                uint8    `json:"decimals"`
	UnderlyingTokenPriceUSD *big.Int `json:"underlyingTokenPriceUSD"`
	ScaledTotalLiquidity    *big.Int `json:"scaledTotalLiquidity"`
	ScaledTotalVariableDebt *big.Int `json:"scaledTotalVariableDebt"`
	AvailableLiquidity      *big.Int `json:"availableLiquidity"`
	TreasuryAmount          *big.Int `json:"treasuryAmount"`

	Configuration               *big.Int       `json:"configuration"`
	LiquidityIndex              *big.Int       `json:"liquidityIndex"`
	CurrentLiquidityRate        *big.Int       `json:"currentLiquidityRate"`
	VariableBorrowIndex         *big.Int       `json:"variableBorrowIndex"`
	CurrentVariableBorrowRate   *big.Int       `json:"currentVariableBorrowRate"`
	CurrentStableBorrowRate     *big.Int       `json:"currentStableBorrowRate"`
	LastUpdateTimestamp         *big.Int       `json:"lastUpdateTimestamp"`
	Id                          uint16         `json:"id"`
	ATokenAddress               common.Address `json:"aTokenAddress"`
	StableDebtTokenAddress      common.Address `json:"stableDebtTokenAddress"`
	VariableDebtTokenAddress    common.Address `json:"variableDebtTokenAddress"`
	InterestRateStrategyAddress common.Address `json:"interestRateStrategyAddress"`
	AccruedToTreasury           *big.Int       `json:"accruedToTreasury"`
	Unbacked                    *big.Int       `json:"unbacked"`
	IsolationModeTotalDebt      *big.Int       `json:"isolationModeTotalDebt"`

	Ltv                    int64 `json:"ltv"`
	LiquidationThreshold   int64 `json:"liquidationThreshold"`
	LiquidationBonus       int64 `json:"liquidationBonus"`
	ReserveFactor          int64 `json:"reserveFactor"`
	BorrowCap              int64 `json:"borrowCap"`
	SupplyCap              int64 `json:"supplyCap"`
	LiquidationProtocolFee int64 `json:"liquidationProtocolFee"`
	EModeCategory          int64 `json:"eModeCategory"`
}

func CollectReservesData(pool *pool.Pool, tokens []AaveToken, blockNumber *big.Int) ([]ReserveData, error) {
	allReservesData := make([]ReserveData, 0)
	for _, token := range tokens {
		fmt.Println(token.UnderlyingAsset)
		data, err := pool.GetReserveData(&bind.CallOpts{BlockNumber: blockNumber}, common.HexToAddress(token.UnderlyingAsset))
		if err != nil {
			return make([]ReserveData, 0), err
		}
		binString := fmt.Sprintf("%b", data.Configuration.Data)
		println(binString)
		println(len(binString))

		var ltv, lt, lb, reserveFactor, borrowCap, supplyCap, liquidationProtocolFee, eModeCategory int64
		if token.UnderlyingAsset != "0x40D16FC0246aD3160Ccc09B8D0D3A2cD28aE6C2f" {
			ltv, _ = strconv.ParseInt(binString[len(binString)-15:], 2, 64)
			lt, _ = strconv.ParseInt(binString[len(binString)-31:len(binString)-16], 2, 64)
			lb, _ = strconv.ParseInt(binString[len(binString)-47:len(binString)-32], 2, 64)
			reserveFactor, _ = strconv.ParseInt(binString[len(binString)-79:len(binString)-64], 2, 64)
			borrowCap, _ = strconv.ParseInt(binString[len(binString)-115:len(binString)-80], 2, 64)
			if len(binString) >= 151 {
				supplyCap, _ = strconv.ParseInt(binString[len(binString)-151:len(binString)-116], 2, 64)
			} else {
				supplyCap = 0
			}
			if len(binString) >= 167 {
				liquidationProtocolFee, _ = strconv.ParseInt(binString[len(binString)-167:len(binString)-152], 2, 64)
			} else {
				liquidationProtocolFee = 0
			}
			if len(binString) >= 175 {
				eModeCategory, _ = strconv.ParseInt(binString[len(binString)-175:len(binString)-168], 2, 64)
			} else {
				eModeCategory = 0
			}
		} else {
			ltv = 0
			lt = 0
			lb = 0
			reserveFactor = 0
			borrowCap = 0
			supplyCap = 0
			liquidationProtocolFee = 0
			eModeCategory = 0
		}

		allReservesData = append(allReservesData, ReserveData{
			BlockNumber:             blockNumber,
			Name:                    token.Name,
			UnderlyingAsset:         token.UnderlyingAsset,
			Decimals:                token.Decimals,
			UnderlyingTokenPriceUSD: token.UnderlyingTokenPriceUSD,
			ScaledTotalLiquidity:    token.ScaledTotalLiquidity,
			ScaledTotalVariableDebt: token.ScaledTotalVariableDebt,
			AvailableLiquidity:      token.AvailableLiquidity,
			TreasuryAmount:          token.TreasuryAmount,

			Configuration:               data.Configuration.Data,
			LiquidityIndex:              data.LiquidityIndex,
			CurrentLiquidityRate:        data.CurrentLiquidityRate,
			VariableBorrowIndex:         data.VariableBorrowIndex,
			CurrentVariableBorrowRate:   data.CurrentVariableBorrowRate,
			CurrentStableBorrowRate:     data.CurrentStableBorrowRate,
			LastUpdateTimestamp:         data.LastUpdateTimestamp,
			Id:                          data.Id,
			ATokenAddress:               data.ATokenAddress,
			StableDebtTokenAddress:      data.StableDebtTokenAddress,
			VariableDebtTokenAddress:    data.VariableDebtTokenAddress,
			InterestRateStrategyAddress: data.InterestRateStrategyAddress,
			AccruedToTreasury:           data.AccruedToTreasury,
			Unbacked:                    data.Unbacked,
			IsolationModeTotalDebt:      data.IsolationModeTotalDebt,

			Ltv:                    ltv,
			LiquidationThreshold:   lt,
			LiquidationBonus:       lb,
			ReserveFactor:          reserveFactor,
			BorrowCap:              borrowCap,
			SupplyCap:              supplyCap,
			LiquidationProtocolFee: liquidationProtocolFee,
			EModeCategory:          eModeCategory,
		})
	}
	return allReservesData, nil
}
