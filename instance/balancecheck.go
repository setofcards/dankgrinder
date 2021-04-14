// Copyright (C) 2021 The Dank Grinder authors.
//
// This source code has been released under the GNU Affero General Public
// License v3.0. A copy of this license is available at
// https://www.gnu.org/licenses/agpl-3.0.en.html

package instance

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dankgrinder/dankgrinder/instance/scheduler"

	"github.com/dankgrinder/dankgrinder/discord"

	"github.com/nickname32/discordhook"
)

func (in *Instance) makebalancerequest(balance string) {
	if in.Features.BalanceCheck.Enable {
		var id, token = in.Features.BalanceCheck.WebhookID, in.Features.BalanceCheck.WebhookToken
		wa, err := discordhook.NewWebhookAPI(id, token, true, nil)
		if err!= nil {
			panic(err)
		}
		msg, err := wa.Execute(nil, &discordhook.WebhookExecuteParams {
			Username: in.Client.User.Username,
			Embeds: []*discordhook.Embed{
				{
					Title: "Balance Check",
					Color: 0x00ff00,
					Description: "current balance: " + balance,
				},
			},
		}, nil, "")

		fmt.Println(msg.ID)
	}
}

func (in *Instance) makeincomerequest(income string) {
	if in.Features.BalanceCheck.Enable {
		var id, token = in.Features.BalanceCheck.WebhookID, in.Features.BalanceCheck.WebhookToken
		wa, err := discordhook.NewWebhookAPI(id, token, true, nil)
		if err!= nil {
			panic(err)
		}

		msg, err := wa.Execute(nil, &discordhook.WebhookExecuteParams {
			Username: in.Client.User.Username,
			Embeds: []*discordhook.Embed{
				{
					Title: "Average Income Check",
					Color: 0x00ff00,
					Description: "average income per hour: " + income,
				},
			},
		}, nil, "")
		fmt.Println(msg.ID)
	}
}

func (in *Instance) balanceCheck(msg discord.Message) {
	if !strings.Contains(msg.Embeds[0].Title, in.Client.User.Username) {
		return
	}
	if !exp.bal.Match([]byte(msg.Embeds[0].Description)) {
		return
	}
	balstr := strings.Replace(exp.bal.FindStringSubmatch(msg.Embeds[0].Description)[1], ",", "", -1)
	balance, err := strconv.Atoi(balstr)
	if err != nil {
		in.Logger.Errorf("error while reading balance: %v", err)
		return
	}
	in.updateBalance(balance)
}

func (in *Instance) updateBalance(balance int) {
	if balance > in.Features.AutoShare.MaximumBalance &&
		in.Features.AutoShare.Enable &&
		in.Master != nil &&
		in != in.Master {
		in.sdlr.PrioritySchedule(&scheduler.Command{
			Value: fmt.Sprintf(
				"pls share %v <@%v>",
				balance-in.Features.AutoShare.MinimumBalance,
				in.Master.Client.User.ID,
			),
			Log: "sharing all balance above minimum with master instance",
		})
	}
	in.balance = balance
	in.lastBalanceUpdate = time.Now()
	in.Logger.Infof(
		"current wallet balance: %v coins",
		numFmt.Sprintf("%d", balance),
	)

	in.makebalancerequest(
		numFmt.Sprintf("%d", balance),
	)
	if in.startingTime.IsZero() {
		in.initialBalance = balance
		in.startingTime = time.Now()
		return
	}

	inc := balance - in.initialBalance
	per := time.Now().Sub(in.startingTime)
	hourlyInc := int(math.Round(float64(inc) / per.Hours()))
	
	in.Logger.Infof(
		"average income: %v coins/h",
		numFmt.Sprintf("%d", hourlyInc),
	)
	in.makeincomerequest(
		numFmt.Sprintf("%d", hourlyInc),
	)
}

