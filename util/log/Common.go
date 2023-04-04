package log

import (
	"os"
	"fmt"
	"log"
    "time"
	"bufio"
    "strings"
	"encoding/json"
	"github.com/barbell-math/block/util/algo"
	customerr "github.com/barbell-math/block/util/err"
)

const LogPartSeparator string="|";

type LogStatus int;
const (
    Error LogStatus=iota
    Warning
    Deprecation
    Info
    Debug
    Invalid
);

func (l LogStatus)String() string {
    switch l {
        case Error: return fmt.Sprintf("Error %s ",LogPartSeparator);
        case Warning: return fmt.Sprintf("Warning %s ",LogPartSeparator);
        case Deprecation: return fmt.Sprintf("Deprecation %s ",LogPartSeparator);
        case Info: return fmt.Sprintf("Info %s ",LogPartSeparator);
        case Debug: return fmt.Sprintf("Debug %s ",LogPartSeparator);
        default: return fmt.Sprintf("Invalid %s ",LogPartSeparator);
    }
}
func LogStatusFromString(s string) (LogStatus,error) {
    switch s {
        case "Error": return Error,nil;
        case "Warning": return Warning,nil;
        case "Deprecation": return Deprecation,nil;
        case "Info": return Info,nil;
        case "Debug": return Debug,nil;
        default: return Invalid,fmt.Errorf("Invalid log status.");
    }
}

type Logger struct {
    file string;
    logFile *os.File;
    logger *log.Logger;
    Log func(message string, val any);
};

func NewLog(status LogStatus, file string) Logger {
    rv:=Logger{ file: file };
    var err error=nil;
    if rv.logFile,err=os.OpenFile(
        file,
        os.O_TRUNC | os.O_CREATE | os.O_WRONLY,
        0644,
    ); err==nil {
        rv.logger=log.New(rv.logFile,status.String(),log.LstdFlags);
    }
    rv.Log=func(message string, val any){
        if b,err:=json.Marshal(val); err==nil {
            rv.logger.Printf(
                "%s %s %s %s",LogPartSeparator,message,LogPartSeparator,b,
            );
        }
    }
    return rv;
}

func NewBlankLog() Logger {
    return Logger{
        Log: func(message string, val any){},
    };
}

func (l *Logger)SetStatus(s LogStatus){
    l.logger.SetPrefix(s.String());
}

func (l *Logger)Close(){
    if l.logFile!=nil {
        l.logFile.Close();
        l.logFile=nil;
    }
}

func (l *Logger)Clear() error {
    if len(l.file)>0 {
        return os.Truncate(l.file,0);
    }
    return LogFileNotSpecified("Nothing to clear.");
}

type LogEntry[T any] struct {
    Status LogStatus;
    Time time.Time;
    Message string;
    Val T;
};
func LogElems[T any](l Logger) algo.Iter[LogEntry[T]] {
    cntr:=0;
    var iterElem T;
    var scanner *bufio.Scanner;
    //f,err:=os.Open(l.file);
    var err error=nil;
    l.logFile,err=os.Open(l.file);
    if err==nil {
        scanner=bufio.NewScanner(l.logFile);
        scanner.Split(bufio.ScanLines);
    }
    return func() (LogEntry[T],error,bool) {
        if err!=nil || !scanner.Scan() {
            l.Close();
            return LogEntry[T]{},err,false;
        }
        cntr++;
        parts:=strings.SplitN(scanner.Text(),LogPartSeparator,4);
        s,serr:=getStatus(parts);
        t,terr:=getTime(parts);
        verr:=getObject(parts,&iterElem);
        var finalErr error=nil;
        if rv:=customerr.AppendError(
            serr,customerr.AppendError(terr,verr),
        ); rv!=nil {
            finalErr=LogLineMalformed(
                fmt.Sprintf("File '%s': Line %d | %s",l.file,cntr,rv),
            );
        }
        return LogEntry[T]{
            Status: s, Time: t, Message: getMessage(parts), Val: iterElem,
        }, finalErr, true;
    }
}

func getStatus(parts []string) (LogStatus,error) {
    if len(parts)>0 {
        return LogStatusFromString(strings.TrimSpace(parts[0]));
    }
    return -1,fmt.Errorf("No log status present.");
}

func getTime(parts []string) (time.Time,error) {
    if len(parts)>=1 {
        if rv,err:=time.Parse(
            "2006/01/02 15:04:05",strings.TrimSpace(parts[1]),
        ); err==nil {
            return rv,err;
        } else {
            return time.Time{},err;
        }
    }
    return time.Time{},fmt.Errorf("No log time present");
}

func getMessage(parts []string) string {
    if len(parts)>=2 {
        return parts[2];
    }
    return "";
}

func getObject[T any](parts []string, elem *T) error {
    if len(parts)>=3 {
        if err:=json.Unmarshal([]byte(parts[3]),elem); err==nil {
            return err;
        } 
        return nil;
    }
    return fmt.Errorf("No object present");
}
