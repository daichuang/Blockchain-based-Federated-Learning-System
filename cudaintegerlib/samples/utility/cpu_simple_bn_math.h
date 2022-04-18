

void add_words(uint32_t *r, uint32_t *x, uint32_t *y, uint32_t count) {
  int     index;
  int64_t sum=0;

  for(index=0;index<count;index++) {
    sum=sum+x[index]+y[index];
    r[index]=sum;
    sum=sum>>32;
  }
}

void sub_words(uint32_t *r, uint32_t *x, uint32_t *y, uint32_t count) {
  int     index;
  int64_t sum=0;

  for(index=0;index<count;index++) {
    sum=sum+x[index]-y[index];
    r[index]=sum;
    sum=sum>>32;
  }
}

