package model

import (
	"fmt"
	"testing"

	"github.com/barbell-math/block/db"
	"github.com/barbell-math/block/settings"
	"github.com/barbell-math/block/util/algo/iter"
	customerr "github.com/barbell-math/block/util/err"
	"github.com/barbell-math/block/util/io/csv"
	logUtil "github.com/barbell-math/block/util/io/log"
	"github.com/barbell-math/block/util/test"
)

var testDB db.DB;

func TestMain(m *testing.M){
    settings.ReadSettings("testData/modelTestSettings.json");
    setup();
    m.Run();
    teardown();
}

func setup(){
    var err error=nil;
    testDB,err=db.NewDB(settings.DBHost(),settings.DBPort(),settings.DBName());
    if err!=nil && err!=db.DataVersionNotAvailable {
        panic("Could not open database for testing.");
    }
    if err=testDB.ResetDB(); err!=nil {
        panic("Could not reset DB for testing. Check location of global init SQL file relative to the ./testData/modelTestSettings.json file.");
    }
    if err:=uploadTestData(); err!=nil {
        panic(fmt.Sprintf(
            "Could not upload data for testing. Check location of testData folder. | %s",
            err,
        ));
    }
}

func setupLogs(debugFile string) (func()) {
    SLIDING_WINDOW_DP_DEBUG=logUtil.NewLog[*dataPoint](logUtil.Debug,
        fmt.Sprintf("%s.dataPoint",debugFile),false,
    );
    SLIDING_WINDOW_MS_DEBUG=logUtil.NewLog[db.ModelState](logUtil.Debug,
        fmt.Sprintf("%s.modelState",debugFile),false,
    );
    return func(){
        SLIDING_WINDOW_DP_DEBUG.Close();
        SLIDING_WINDOW_MS_DEBUG.Close();
    }
}

func uploadTestData() error {
    return customerr.ChainedErrorOps(
        func(r ...any) (any,error) {
            return db.Create(&testDB,db.Client{
                Id: 1,
                FirstName: "testF",
                LastName: "testL",
                Email: "test@test.com",
            });
        }, func(r ...any) (any,error) {
            return nil,csv.CSVToStruct[db.StateGenerator](csv.CSVFileSplitter(
                "../../data/testData/StateGeneratorTestData.csv",',','#',
            ),"1/2/2006").ForEach(
            func(index int, val db.StateGenerator) (iter.IteratorFeedback, error) {
                db.Create(&testDB,val);
                return iter.Continue,nil;
            });
            //return nil,csv.CSVToStruct(
            //    "../../data/testData/StateGeneratorTestData.csv",',',"",
            //    func(s *db.StateGenerator){
            //        db.Create(&testDB,*s);
            //});
        }, func(r ...any) (any,error) {
            return nil,csv.CSVToStruct[db.ExerciseType](csv.CSVFileSplitter(
                "../../data/testData/ExerciseTypeTestData.csv",',','#',
            ),"1/2/2006").ForEach(
            func(index int, val db.ExerciseType) (iter.IteratorFeedback, error) {
                db.Create(&testDB,val);
                return iter.Continue,nil;
            });
            //return nil,csv.CSVToStruct(
            //    "../../data/testData/ExerciseTypeTestData.csv",',',"",
            //    func(e *db.ExerciseType){
            //        db.Create(&testDB,*e);
            //});
        },func(r ...any) (any,error) {
            return nil,csv.CSVToStruct[db.ExerciseFocus](csv.CSVFileSplitter(
                "../../data/testData/ExerciseFocusTestData.csv",',','#',
            ),"1/2/2006").ForEach(
            func(index int, val db.ExerciseFocus) (iter.IteratorFeedback, error) {
                db.Create(&testDB,val);
                return iter.Continue,nil;
            });
            //return nil,csv.CSVToStruct(
            //    "../../data/testData/ExerciseFocusTestData.csv",',',"",
            //    func(e *db.ExerciseFocus){
            //        db.Create(&testDB,*e);
            //});
        },func(r ...any) (any,error) {
            return nil,csv.CSVToStruct[db.Exercise](csv.CSVFileSplitter(
                "../../data/testData/ExerciseTestData.csv",',','#',
            ),"1/2/2006").ForEach(
            func(index int, val db.Exercise) (iter.IteratorFeedback, error) {
                db.Create(&testDB,val);
                return iter.Continue,nil;
            });
            //return nil,csv.CSVToStruct(
            //    "../../data/testData/ExerciseTestData.csv",',',"",
            //    func(e *db.Exercise){
            //        db.Create(&testDB,*e);
            //});
        },func(r ...any) (any,error) {
            return nil,csv.CSVToStruct[db.Rotation](csv.CSVFileSplitter(
                "../../data/testData/RotationTestData.csv",',','#',
            ),"1/2/2006").ForEach(
            func(index int, val db.Rotation) (iter.IteratorFeedback, error) {
                db.Create(&testDB,val);
                return iter.Continue,nil;
            });
            //return nil,csv.CSVToStruct(
            //    "../../data/testData/RotationTestData.csv",',',"1/2/2006",
            //    func(r *db.Rotation){
            //        db.Create(&testDB,*r);
            //});
        },func(r ...any) (any,error) {
            return nil,csv.CSVToStruct[db.TrainingLog](csv.CSVFileSplitter(
                "../../data/testData/AugmentedTrainingLogTestData.csv",',','#',
            ),"1/2/2006").ForEach(
            func(index int, val db.TrainingLog) (iter.IteratorFeedback, error) {
                db.Create(&testDB,val);
                return iter.Continue,nil;
            });
            //return nil,csv.CSVToStruct(
            //    "../../data/testData/AugmentedTrainingLogTestData.csv",',',"1/2/2006",
            //    func(t *db.TrainingLog){
            //        db.Create(&testDB,*t);
            //});
    });
}

func teardown(){
    //testDB.ResetDB();
    testDB.Close();
}

func TestModelCreation(t *testing.T){
    cntr:=0;
    m:=fatigueAwareModel();
    m.IterLHS(func(r int, c int, v float64){
        cntr++;
    });
    test.BasicTest(49,cntr,"LHS Lin reg wrong size for model.",t);
    cntr=0;
    m.IterRHS(func(r int, c int, v float64){
        cntr++;
    });
    test.BasicTest(7,cntr,"RHS Lin reg wrong size for model.",t);
}

func TestIntensityPrediction(t *testing.T){
    ms:=db.ModelState{
        Eps: 0, Eps1: 0, Eps2: 0, Eps3: 0, Eps4: 0,
        Eps5: 0, Eps6: 0, Eps7: 0,
    };
    tl:=db.TrainingLog{
        Weight: 0, Sets: 0, Reps: 0, Intensity: 0, Effort: 0,
        InterWorkoutFatigue: 0, InterExerciseFatigue: 0,
    };
    res:=IntensityPrediction(&ms,&tl);
    test.BasicTest(float64(0.0),res,
        "Intensity prediction produced incorrect value.",t,
    );
}

func TestEffortPrediction(t *testing.T){
    //Eps1 has to be 1 to avoid div by 0 error
    ms:=db.ModelState{
        Eps: 0, Eps1: 1, Eps2: 0, Eps3: 0,
        Eps4: 0, Eps5: 0, Eps6: 0, Eps7: 0,
    };
    tl:=db.TrainingLog{
        Weight: 0, Sets: 0, Reps: 0, Intensity: 0, Effort: 0,
    };
    res:=EffortPrediction(&ms,&tl);
    test.BasicTest(float64(0.0),res,
        "Effort prediction produced incorrect value.",t,
    );
}

func TestLatentFatiguePrediction(t *testing.T){
    //Eps2 has to be 1 to avoid div by 0 error
    ms:=db.ModelState{
        Eps: 0, Eps1: 0, Eps2: 1, Eps3: 0,
        Eps4: 0, Eps5: 0, Eps6: 0, Eps7: 0,
    };
    tl:=db.TrainingLog{
        Weight: 0, Sets: 0, Reps: 0, Intensity: 0, Effort: 0,
    };
    res:=LatentFatiguePrediction(&ms,&tl);
    test.BasicTest(float64(0.0),res,
        "Latent fatigue prediction produced incorrect value.",t,
    );
}

func TestInterWorkoutFatiguePrediction(t *testing.T){
    //Eps3 has to be 1 to avoid div by 0 error
    ms:=db.ModelState{
        Eps: 0, Eps1: 0, Eps2: 0, Eps3: 1,
        Eps4: 0, Eps5: 0, Eps6: 0, Eps7: 0,
    };
    tl:=db.TrainingLog{
        Weight: 0, Sets: 0, Reps: 0, Intensity: 0, Effort: 0,
    };
    res:=InterWorkoutFatiguePrediciton(&ms,&tl);
    test.BasicTest(float64(0.0),res,
        "Inter workout fatigue prediction produced incorrect value.",t,
    );
}

func TestInterExerciseFatiguePrediction(t *testing.T){
    //Eps4 has to be 1 to avoid div by 0 error
    ms:=db.ModelState{
        Eps: 0, Eps1: 0, Eps2: 0, Eps3: 0,
        Eps4: 1, Eps5: 0, Eps6: 0, Eps7: 0,
    };
    tl:=db.TrainingLog{
        Weight: 0, Sets: 0, Reps: 0, Intensity: 0, Effort: 0,
    };
    res:=InterExerciseFatiguePrediction(&ms,&tl);
    test.BasicTest(float64(0.0),res,
        "Inter exercise fatigue prediction produced incorrect value.",t,
    );
}

func TestSetsPrediction(t *testing.T){
    //Eps6 has to be 1 to avoid div by 0 error
    ms:=db.ModelState{
        Eps: 0, Eps1: 0, Eps2: 0, Eps3: 0,
        Eps4: 0, Eps5: 0, Eps6: 1, Eps7: 0,
    };
    tl:=db.TrainingLog{
        Weight: 0, Sets: 0, Reps: 0, Intensity: 0, Effort: 0,
    };
    res:=SetsPrediction(&ms,&tl);
    test.BasicTest(float64(1.0),res,
        "Sets prediction produced incorrect value.",t,
    );
}

func TestRepsPrediction(t *testing.T){
    //Eps7 has to be 1 to avoid div by 0 error
    ms:=db.ModelState{
        Eps: 0, Eps1: 0, Eps2: 0, Eps3: 0,
        Eps4: 0, Eps5: 0, Eps6: 0, Eps7: 1,
    };
    tl:=db.TrainingLog{
        Weight: 0, Sets: 0, Reps: 0, Intensity: 0, Effort: 0,
    };
    res:=RepsPrediction(&ms,&tl);
    test.BasicTest(float64(1.0),res,
        "Reps prediction produced incorrect value.",t,
    );
}
