package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/barbell-math/block/db"
	"github.com/barbell-math/block/util/algo/iter"
	"github.com/barbell-math/block/util/dataStruct"
	"github.com/barbell-math/block/util/io/log"
	"github.com/barbell-math/block/util/test"
)

func invalidCheck(slidingWindowSg SlidingWindowStateGen, err error) (func(t *testing.T)){
    return func(t *testing.T){
        if !IsInvalidPredictionState(err) {
            test.FormatError(InvalidPredictionState(""),err,
                "The wrong error was raised when creating an invalid prediction generator.",t,
            );
        }
    }
}
func TestNewSlidingWindowStateGenInvalidTimeFrameLimits(t *testing.T){
    invalidCheck(NewSlidingWindowStateGen(
        dataStruct.Pair[int,int]{A: 1,B: 0},dataStruct.Pair[int,int]{A: 0, B: 1},1,
    ))(t);
}
func TestNewSlidingWindowStateGenInvalidWindowLimits(t *testing.T){
    invalidCheck(NewSlidingWindowStateGen(
        dataStruct.Pair[int,int]{A: 0, B: 1},dataStruct.Pair[int,int]{A: 1, B: 0},1,
    ))(t);
}
func TestNewSlidingWindowStateGenInvalidWindowSize(t *testing.T){
    invalidCheck(NewSlidingWindowStateGen(
        dataStruct.Pair[int,int]{A: 0, B: 1},dataStruct.Pair[int,int]{A: 0, B: 2},1,
    ))(t);
    invalidCheck(NewSlidingWindowStateGen(
        dataStruct.Pair[int,int]{A: 1, B: 2},dataStruct.Pair[int,int]{A: 0, B: 2},1,
    ))(t);
}
func TestNewSlidingWindowValid(t *testing.T){
    _,err:=NewSlidingWindowStateGen(
        dataStruct.Pair[int,int]{A: 0, B: 1},dataStruct.Pair[int,int]{A: 0, B: 1},1,
    );
    test.BasicTest(nil,err,
        "Creating a sliding window resulted in an error when it shouldn't have.",t,
    );
}

func TestNewSlidingWindowConstrainedThreadAllocation(t *testing.T){
    sw,err:=NewSlidingWindowStateGen(
        dataStruct.Pair[int,int]{A: 0, B: 1},dataStruct.Pair[int,int]{A: 0, B: 1},0,
    );
    test.BasicTest(nil,err,
        "Creating a sliding window resulted in an error when it shouldn't have.",t,
    );
    test.BasicTest(1,sw.allotedThreads,
        "The sliding window was allotted the wrong number of threads.",t,
    );
}

//func TestGenerateModelStates(t *testing.T){
//    setLogs("./debugLogs/SlidingWindow.log");
//    ch:=make(chan<- []error);
//    sw,_:=NewSlidingWindowStateGen(
//        dataStruct.Pair[int]{0, 1},dataStruct.Pair[int]{0, 1},0,
//    );
//    sw.GenerateClientModelStates(&testDB,db.Client{ Id: 1 },ch);
//}

func TestGenerateModelState(t *testing.T){
    tmp:=setupLogs("./debugLogs/SlidingWindowStateGeneratorGood");
    baseTime,_:=time.Parse("01/02/2006","09/10/2022");
    ch:=make(chan<- StateGeneratorRes);
    timeFrame:=dataStruct.Pair[int,int]{A: 0, B: 500};
    window:=dataStruct.Pair[int,int]{A: 0, B: 10};
    sw,_:=NewSlidingWindowStateGen(timeFrame,window,1);
        //dataStruct.Pair[int,int]{4, 500},dataStruct.Pair[int,int]{5, 10},0,
        //dataStruct.Pair[int,int]{0, 500},dataStruct.Pair[int,int]{0, 1},0,
    missingData:=missingModelStateData{
        ClientID: 1,
        ExerciseID: 15,
        Date: baseTime,
    };
    err:=sw.GenerateModelState(&testDB,missingData,ch);
    fmt.Println("ERR: ",err);
    tmp();
    runModelStateDebugLogTests(baseTime,
        missingData.ClientID,missingData.ExerciseID,int(SlidingWindowStateGenId),
        timeFrame,window,t,
    );
    runDataPointDebugLogTests(t);
}

func runDataPointDebugLogTests(t *testing.T){
    initialDate:=time.Time{};
    log.LogElems(SLIDING_WINDOW_DP_DEBUG).Filter(
    func(index int, val log.LogEntry[*dataPoint]) bool {
        if index==0 {
            initialDate=val.Val.DatePerformed;
            return false;
        }
        return true;
    }).ForEach(
    func(index int, val log.LogEntry[*dataPoint]) (iter.IteratorFeedback, error) {
        test.BasicTest(true,initialDate.Sub(val.Val.DatePerformed)>=0,
            "Training log dates did not continually decrease from query.",t,
        );
        initialDate=val.Val.DatePerformed;
        return iter.Continue,nil;
    });
}

func runModelStateDebugLogTests(baseTime time.Time, cId int, eId int, sId int,
        timeFrame dataStruct.Pair[int,int],
        window dataStruct.Pair[int,int], t *testing.T){
    initialMse:=0.0;
    log.LogElems(SLIDING_WINDOW_MS_DEBUG).Next(
    func(index int,
        val log.LogEntry[db.ModelState],
        status iter.IteratorFeedback,
    ) (iter.IteratorFeedback, log.LogEntry[db.ModelState], error) {
        if status!=iter.Break {
            test.BasicTest(true,val.Val.TimeFrame>=timeFrame.A,
                "A model state had a time frame less than the selected lowest value.",t,
            );
            test.BasicTest(true,val.Val.TimeFrame<=timeFrame.B,
                "A model state had a time frame greater than the selected highest value.",t,
            );
            test.BasicTest(true,val.Val.Win>=window.A,
                "A model state had a window less than the selected lowest value.",t,
            );
            test.BasicTest(true,val.Val.Win<=window.B,
                "A model state had a window greater than the selected highest value.",t,
            );
            test.BasicTest(cId,val.Val.ClientID,
                "A model state had the incorrect client ID.",t,
            );
            test.BasicTest(eId,val.Val.ExerciseID,
                "A model state had the incorrect client ID.",t,
            );
            test.BasicTest(sId,val.Val.StateGeneratorID,
                "A model state had the incorrect state generator ID.",t,
            );
            y1,m1,d1:=val.Val.Date.Date();
            y2,m2,d2:=baseTime.Date();
            test.BasicTest(y2,y1,"A model state had an incorrect year.",t);
            test.BasicTest(m2,m1,"A model state had an incorrect month.",t);
            test.BasicTest(d2,d1,"A model state had an incorrect day.",t);
        }
        return iter.Continue,val,nil;
    }).Filter(func(index int, val log.LogEntry[db.ModelState]) bool {
        if index==0 {
            initialMse=val.Val.Mse;
            return false;
        }
        return true;
    }).ForEach(
    func(index int, val log.LogEntry[db.ModelState]) (iter.IteratorFeedback, error) {
        test.BasicTest(true,initialMse>val.Val.Mse,
            "Mse values did not continually decrease.",t,
        );
        initialMse=val.Val.Mse;
        return iter.Continue,nil;
    });
}
