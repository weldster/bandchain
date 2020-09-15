package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/bandchain/chain/emitter/common"
	"github.com/bandprotocol/bandchain/chain/x/oracle"
	"github.com/bandprotocol/bandchain/chain/x/oracle/types"
)

func (app *App) emitOracleModule() {
	dataSources := app.OracleKeeper.GetAllDataSources(app.DeliverContext)
	for idx, ds := range dataSources {
		id := types.DataSourceID(idx + 1)
		app.emitSetDataSource(id, ds, nil)
		common.EmitNewDataSourceRequest(app, id)
	}
	oracleScripts := app.OracleKeeper.GetAllOracleScripts(app.DeliverContext)
	for idx, os := range oracleScripts {
		id := types.OracleScriptID(idx + 1)
		app.emitSetOracleScript(id, os, nil)
		common.EmitNewOracleScriptRequest(app, id)
	}
	rqCount := app.OracleKeeper.GetRequestCount(app.DeliverContext)
	for rid := types.RequestID(1); rid <= types.RequestID(rqCount); rid++ {
		req := app.OracleKeeper.MustGetRequest(app.DeliverContext, rid)
		app.Write("NEW_REQUEST", common.JsDict{
			"id":               rid,
			"oracle_script_id": req.OracleScriptID,
			"calldata":         common.ParseBytes(req.Calldata),
			"ask_count":        len(req.RequestedValidators),
			"min_count":        req.MinCount,
			"tx_hash":          nil,
			"client_id":        req.ClientID,
			"resolve_status":   types.ResolveStatus_Open,
		})
		if app.OracleKeeper.HasResult(app.DeliverContext, rid) {
			app.emitUpdateResult(rid)
		}
		app.emitRawRequestAndValRequest(rid, req)
		reps := app.OracleKeeper.GetReports(app.DeliverContext, rid)
		for _, rep := range reps {
			app.emitReportAndRawReport(nil, rid, rep.Validator, nil, rep.RawReports)
		}
	}
}

func (app *App) emitSetDataSource(id types.DataSourceID, ds types.DataSource, txHash []byte) {
	app.Write("SET_DATA_SOURCE", common.JsDict{
		"id":          id,
		"name":        ds.Name,
		"description": ds.Description,
		"owner":       ds.Owner.String(),
		"executable":  app.OracleKeeper.GetFile(ds.Filename),
		"tx_hash":     txHash,
	})
}

func (app *App) emitSetOracleScript(id types.OracleScriptID, os types.OracleScript, txHash []byte) {
	app.Write("SET_ORACLE_SCRIPT", common.JsDict{
		"id":              id,
		"name":            os.Name,
		"description":     os.Description,
		"owner":           os.Owner.String(),
		"schema":          os.Schema,
		"codehash":        os.Filename,
		"source_code_url": os.SourceCodeURL,
		"tx_hash":         txHash,
	})
}

func (app *App) emitHistoricalValidatorStatus(operatorAddress sdk.ValAddress) {
	status := app.OracleKeeper.GetValidatorStatus(app.DeliverContext, operatorAddress).IsActive
	common.EmitHistoricalValidatorStatus(app, operatorAddress, status, app.DeliverContext.BlockTime().UnixNano())
}

func (app *App) emitRawRequestAndValRequest(requestID types.RequestID, req types.Request) {
	for _, raw := range req.RawRequests {
		app.Write("NEW_RAW_REQUEST", common.JsDict{
			"request_id":     requestID,
			"external_id":    raw.ExternalID,
			"data_source_id": raw.DataSourceID,
			"calldata":       common.ParseBytes(raw.Calldata),
		})
	}
	for _, val := range req.RequestedValidators {
		app.Write("NEW_VAL_REQUEST", common.JsDict{
			"request_id": requestID,
			"validator":  val.String(),
		})
	}
}

func (app *App) emitReportAndRawReport(
	txHash []byte, rid types.RequestID, validator sdk.ValAddress, reporter sdk.AccAddress, rawReports []types.RawReport,
) {
	app.Write("NEW_REPORT", common.JsDict{
		"tx_hash":    txHash,
		"request_id": rid,
		"validator":  validator.String(),
		"reporter":   reporter.String(),
	})
	for _, data := range rawReports {
		app.Write("NEW_RAW_REPORT", common.JsDict{
			"request_id":  rid,
			"validator":   validator.String(),
			"external_id": data.ExternalID,
			"data":        common.ParseBytes(data.Data),
			"exit_code":   data.ExitCode,
		})
	}
}

func (app *App) emitUpdateResult(id types.RequestID) {
	result := app.OracleKeeper.MustGetResult(app.DeliverContext, id)
	app.Write("UPDATE_REQUEST", common.JsDict{
		"id":             id,
		"request_time":   result.ResponsePacketData.RequestTime,
		"resolve_time":   result.ResponsePacketData.ResolveTime,
		"resolve_status": result.ResponsePacketData.ResolveStatus,
		"result":         common.ParseBytes(result.ResponsePacketData.Result),
	})
}

// handleMsgRequestData implements emitter handler for MsgRequestData.
func (app *App) handleMsgRequestData(
	txHash []byte, msg oracle.MsgRequestData, evMap common.EvMap, extra common.JsDict,
) {
	id := types.RequestID(common.Atoi(evMap[types.EventTypeRequest+"."+types.AttributeKeyID][0]))
	req := app.OracleKeeper.MustGetRequest(app.DeliverContext, id)
	app.Write("NEW_REQUEST", common.JsDict{
		"id":               id,
		"tx_hash":          txHash,
		"oracle_script_id": msg.OracleScriptID,
		"calldata":         common.ParseBytes(msg.Calldata),
		"ask_count":        msg.AskCount,
		"min_count":        msg.MinCount,
		"sender":           msg.Sender.String(),
		"client_id":        msg.ClientID,
		"resolve_status":   types.ResolveStatus_Open,
	})
	app.emitRawRequestAndValRequest(id, req)
	common.EmitSetRequestCountPerDay(app, app.DeliverContext.BlockTime().UnixNano())
	common.EmitUpdateOracleScriptRequest(app, msg.OracleScriptID)
	for _, raw := range req.RawRequests {
		common.EmitUpdateDataSourceRequest(app, raw.DataSourceID)
		common.EmitUpdateRelatedDsOs(app, raw.DataSourceID, msg.OracleScriptID)
	}
	os := app.OracleKeeper.MustGetOracleScript(app.DeliverContext, msg.OracleScriptID)
	extra["id"] = id
	extra["name"] = os.Name
	extra["schema"] = os.Schema
}

// handleMsgReportData implements emitter handler for MsgReportData.
func (app *App) handleMsgReportData(
	txHash []byte, msg oracle.MsgReportData, evMap common.EvMap, extra common.JsDict,
) {
	app.emitReportAndRawReport(txHash, msg.RequestID, msg.Validator, msg.Reporter, msg.RawReports)
}

// handleMsgCreateDataSource implements emitter handler for MsgCreateDataSource.
func (app *App) handleMsgCreateDataSource(
	txHash []byte, msg oracle.MsgCreateDataSource, evMap common.EvMap, extra common.JsDict,
) {
	id := types.DataSourceID(common.Atoi(evMap[types.EventTypeCreateDataSource+"."+types.AttributeKeyID][0]))
	ds := app.BandApp.OracleKeeper.MustGetDataSource(app.DeliverContext, id)
	app.emitSetDataSource(id, ds, txHash)
	common.EmitNewDataSourceRequest(app, id)
	extra["id"] = id
}

// handleMsgCreateOracleScript implements emitter handler for MsgCreateOracleScript.
func (app *App) handleMsgCreateOracleScript(
	txHash []byte, msg oracle.MsgCreateOracleScript, evMap common.EvMap, extra common.JsDict,
) {
	id := types.OracleScriptID(common.Atoi(evMap[types.EventTypeCreateOracleScript+"."+types.AttributeKeyID][0]))
	os := app.BandApp.OracleKeeper.MustGetOracleScript(app.DeliverContext, id)
	app.emitSetOracleScript(id, os, txHash)
	common.EmitNewOracleScriptRequest(app, id)
	extra["id"] = id
}

// handleMsgEditDataSource implements emitter handler for MsgEditDataSource.
func (app *App) handleMsgEditDataSource(
	txHash []byte, msg oracle.MsgEditDataSource, evMap common.EvMap, extra common.JsDict,
) {
	id := msg.DataSourceID
	ds := app.BandApp.OracleKeeper.MustGetDataSource(app.DeliverContext, id)
	app.emitSetDataSource(id, ds, txHash)
}

// handleMsgEditOracleScript implements emitter handler for MsgEditOracleScript.
func (app *App) handleMsgEditOracleScript(
	txHash []byte, msg oracle.MsgEditOracleScript, evMap common.EvMap, extra common.JsDict,
) {
	id := msg.OracleScriptID
	os := app.BandApp.OracleKeeper.MustGetOracleScript(app.DeliverContext, id)
	app.emitSetOracleScript(id, os, txHash)
}

// handleEventRequestExecute implements emitter handler for EventRequestExecute.
func (app *App) handleEventRequestExecute(evMap common.EvMap) {
	app.emitUpdateResult(types.RequestID(common.Atoi(evMap[types.EventTypeResolve+"."+types.AttributeKeyID][0])))
}

// handleMsgAddReporter implements emitter handler for MsgAddReporter.
func (app *App) handleMsgAddReporter(
	txHash []byte, msg oracle.MsgAddReporter, evMap common.EvMap, extra common.JsDict,
) {
	val, _ := app.StakingKeeper.GetValidator(app.DeliverContext, msg.Validator)
	extra["validator_moniker"] = val.GetMoniker()
	app.AddAccountsInTx(msg.Reporter)
	app.Write("SET_REPORTER", common.JsDict{
		"reporter":  msg.Reporter,
		"validator": msg.Validator,
	})
}

// handleMsgRemoveReporter implements emitter handler for MsgRemoveReporter.
func (app *App) handleMsgRemoveReporter(
	txHash []byte, msg oracle.MsgRemoveReporter, evMap common.EvMap, extra common.JsDict,
) {
	val, _ := app.StakingKeeper.GetValidator(app.DeliverContext, msg.Validator)
	extra["validator_moniker"] = val.GetMoniker()
	app.AddAccountsInTx(msg.Reporter)
	app.Write("REMOVE_REPORTER", common.JsDict{
		"reporter":  msg.Reporter,
		"validator": msg.Validator,
	})
}

// handleMsgActivate implements emitter handler for handleMsgActivate.
func (app *App) handleMsgActivate(
	txHash []byte, msg oracle.MsgActivate, evMap common.EvMap, extra common.JsDict,
) {
	app.emitUpdateValidatorStatus(msg.Validator)
	app.emitHistoricalValidatorStatus(msg.Validator)
}

// handleEventDeactivate implements emitter handler for EventDeactivate.
func (app *App) handleEventDeactivate(evMap common.EvMap) {
	addr, _ := sdk.ValAddressFromBech32(evMap[types.EventTypeDeactivate+"."+types.AttributeKeyValidator][0])
	app.emitUpdateValidatorStatus(addr)
	app.emitHistoricalValidatorStatus(addr)
}