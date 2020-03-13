module Styles = {
  open Css;
  let fullWidth = style([width(`percent(100.0)), display(`flex)]);
  let container = style([width(`px(68))]);
  let hashContainer = style([maxWidth(`px(140))]);
  let paddingTopContainer = style([paddingTop(`px(5))]);
  let statusContainer =
    style([maxWidth(`px(95)), display(`flex), flexDirection(`row), alignItems(`center)]);
  let logo = style([width(`px(20)), marginLeft(`auto), marginRight(`px(15))]);
};
[@react.component]
let make = (~txs: list(TxHook.Tx.t)) => {
  <>
    <THead>
      <Row>
        <HSpacing size={`px(20)} />
        <Col size=1.67>
          <div className=Styles.fullWidth>
            <Text value="TX HASH" size=Text.Sm weight=Text.Bold color=Colors.mediumLightGray />
          </div>
        </Col>
        <Col size=1.05>
          <div className=Styles.fullWidth>
            <AutoSpacing dir="left" />
            <Text
              value="GAS FEE (BAND)"
              size=Text.Sm
              weight=Text.Bold
              color=Colors.mediumLightGray
            />
            <HSpacing size={`px(20)} />
          </div>
        </Col>
        <Col> <div className=Styles.container /> </Col>
        <Col size=5.>
          <div className=Styles.fullWidth>
            <Text value="ACTIONS" size=Text.Sm weight=Text.Bold color=Colors.mediumLightGray />
          </div>
        </Col>
        <HSpacing size={`px(20)} />
      </Row>
    </THead>
    {txs
     ->Belt.List.map(({blockHeight, hash, timestamp, fee, gasUsed, messages, sender, success}) => {
         <TBody key={hash |> Hash.toHex}>
           <Row>
             <HSpacing size={`px(20)} />
             <Col size=1.67 alignSelf=Col.Start>
               <div className={Css.merge([Styles.hashContainer, Styles.paddingTopContainer])}>
                 <Text
                   block=true
                   code=true
                   spacing={Text.Em(0.02)}
                   value={hash |> Hash.toHex(~upper=true)}
                   weight=Text.Medium
                   ellipsis=true
                 />
               </div>
             </Col>
             <Col size=1.05 alignSelf=Col.Start>
               <div className={Css.merge([Styles.fullWidth, Styles.paddingTopContainer])}>
                 <AutoSpacing dir="left" />
                 <Text
                   block=true
                   code=true
                   spacing={Text.Em(0.02)}
                   value={fee->TxHook.Coin.getBandAmountFromCoins->Format.fPretty}
                   weight=Text.Medium
                   ellipsis=true
                 />
                 <HSpacing size={`px(20)} />
               </div>
             </Col>
             <Col> <div className=Styles.container /> </Col>
             <Col size=5. alignSelf=Col.Start>
               {messages
                ->Belt.List.map(msg => {
                    <> <Msg msg width=330 success /> <VSpacing size=Spacing.md /> </>
                  })
                ->Belt.List.toArray
                ->React.array}
             </Col>
             <HSpacing size={`px(20)} />
           </Row>
         </TBody>
       })
     ->Array.of_list
     ->React.array}
  </>;
};