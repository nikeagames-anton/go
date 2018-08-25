package backup

import (
    "archive/zip"
    "errors"
    "io"
    "io/ioutil"
    "os"
)

var (
    errInvalidPath = errors.New( "backup: invalid path" )
)

type Backup struct{
    zw *zip.Writer
    f  *os.File
}

func New( path string ) ( *Backup, error ) {
    
    if fileExists( path ) {
        return nil, errInvalidPath
    }
    
    var (
        err error
        b   *Backup
    )
    
    b        = new( Backup )
    b.f, err = os.OpenFile( path, os.O_CREATE | os.O_WRONLY, 0644 )
    if err != nil {
        return nil, err
    }
    
    b.zw = zip.NewWriter( b.f )
    if err != nil {
        b.f.Close()
        return nil, err
    }
    
    return b, nil
    
}

// Copy
//
// - Copies entries in the from directory to to directory in the archive
// - from can be either absolute path or relative path
// - to must be relative path in the archive
//
// @return int64 - the size written
// @return error - error

func ( b *Backup ) Copy( from string, to string ) ( int, error ) {
    
    if !fileExists( from ) {
        return 0, errInvalidPath
    }
    
    fi, err := os.Stat( from )
    if err != nil {
        return 0, err
    }
    
    if fi.IsDir() {
        return b.copyRecursive( from, to )
    } else {
        return b.copyFile( from, to )
    }
    
}

func ( b *Backup ) copyRecursive( from string, to string ) ( int, error ) {
    
    files, err := ioutil.ReadDir( from )
    if err != nil {
        return 0, err
    }
    
    from = toDirPath( from )
    to   = toDirPath( to )
    
    if to != "" {
        err  = b.NewDir( to )
        if err != nil {
            return 0, err
        }
    }
    
    var (
        written int = 0
    )
    
    for _, file := range files {
        realPath := from + file.Name()
        zipPath  := to + file.Name()
        
        var (
            wsz int
            err error
        )
        
        if file.IsDir() {
            wsz, err = b.copyRecursive( realPath, zipPath )
        } else {
            wsz, err = b.copyFile( realPath, zipPath )
        }
        
        written += wsz
        if err != nil {
            return written, err
        }
    }
    
    return written, nil
    
}

func ( b *Backup ) copyFile( from string, to string ) ( int, error ) {
    buf, err := ioutil.ReadFile( from )
    if err != nil {
        return 0, err
    }

    w, err := b.NewFile( to )
    if err != nil {
        return 0, err
    }

    written, err := w.Write( buf )
    if err != nil {
        return written, err
    }
    return written, nil
}

func ( b *Backup ) NewFile( path string ) ( io.Writer, error ) {
    path = toFilePath( path )
    return b.zw.Create( path )
}

func ( b *Backup ) NewDir( path string ) error {
    path    = toDirPath( path )
    _, err := b.zw.Create( path )
    if err != nil {
        return err
    }
    return nil
}

func ( b *Backup ) Close() error {
    if err := b.zw.Close(); err != nil {
        return err
    }
    return b.f.Close()
}

func cleanPath( path string ) string {
    if len( path ) == 0 {
        return path
    }
    if path[:1] == "/" {
        return path[1:]
    }
    return path
}

func toFilePath( path string ) string {
    path = cleanPath( path )
    if path[len( path ) - 1:] == "/" {
        path = path[:len( path ) - 1]
    }
    return path
}

func toDirPath( path string ) string {
    path = cleanPath( path )
    if path == "" {
        return path
    }
    if path[len( path ) - 1:] != "/" {
        path = path + "/"
    }
    return path
}

func fileExists( path string ) bool {
    _, err := os.Stat( path )
    return !os.IsNotExist( err )
}
