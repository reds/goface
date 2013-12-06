#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include <opencv/cv.h>
#include <opencv/highgui.h>  // loadimage

CvHaarClassifierCascade *default_frontalface_cascade;
CvMemStorage* tmp;

static int init( char* ffc ) {
  default_frontalface_cascade = cvLoad ( ffc, 0, 0, 0 );
  tmp = cvCreateMemStorage ( 0 );

  return 1;
}

static char* process_image ( char* fn ) {
  printf ( "processing %s\n", fn );
  IplImage* img = cvLoadImage ( fn, CV_LOAD_IMAGE_COLOR );
  if ( img == NULL ) {
    printf ( "img NULL\n" );
    return "";
  }
  CvSeq* faces = cvHaarDetectObjects ( img,
				       default_frontalface_cascade,
				       tmp,
				       1.1,
				       3,
				       CV_HAAR_DO_CANNY_PRUNING,
				       cvSize ( 0, 0 ) );
  char* f = malloc(1);
  f[0] = '\0';
  int i;
  for ( i = 0; i < (faces ? faces->total : 0); i++ ) {
    CvRect* r = (CvRect*) cvGetSeqElem ( faces, i );
    char ff[1024];
    int n = snprintf ( ff, sizeof ff, "%d %d %d %d\n", r->x, r->y, r->width, r->height );
    f = realloc ( f, strlen(f) + n + 1 );
    strcat ( f, ff );
  }
  cvReleaseImage ( &img );
  return f;
}

#ifdef TEST

int main (int argc, char* argv[] ) {
  init("haarcascade_frontalface_default.xml");
  int i;
  for ( i = 1; i < argc; i++ ) {
    printf ( "%d %s\n", i, argv[i] );
    process_image ( argv[i] );
  }
}

#endif

