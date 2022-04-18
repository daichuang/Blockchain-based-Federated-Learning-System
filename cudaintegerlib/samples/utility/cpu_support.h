
#ifdef __GMP_H__

void to_mpz(mpz_t r, uint32_t *x, uint32_t count) {
  mpz_import(r, count, -1, sizeof(uint32_t), 0, 0, x);
}

void from_mpz(mpz_t s, uint32_t *x, uint32_t count) {
  size_t words;

  if(mpz_sizeinbase(s, 2)>count*32) {
    fprintf(stderr, "from_mpz failed -- result does not fit\n");
    exit(1);
  }

  mpz_export(x, &words, -1, sizeof(uint32_t), 0, 0, s);
  while(words<count)
    x[words++]=0;
}

#endif

uint32_t random_word() {
  uint32_t random;

  random=rand() & 0xFFFF;
  random=(random<<16) + (rand() & 0xFFFF);
  return random;
}

void zero_words(uint32_t *x, uint32_t count) {
  int index;

  for(index=0;index<count;index++)
    x[index]=0;
}

void print_words(uint32_t *x, uint32_t count) {
  int index;

  for(index=count-1;index>=0;index--)
    printf("%08X", x[index]);
  printf("\n");
}

void copy_words(uint32_t *from, uint32_t *to, uint32_t count) {
  int index;

  for(index=0;index<count;index++)
    to[index]=from[index];
}

int32_t nibble(char c) {
  if('0'<=c && c<='9')
    return c-'0';
  else if('a'<=c && c<='f')
    return c-'a'+10;
  else if('A'<=c && c<='F')
    return c-'A'+10;
  else {
    printf("Invalid nibble: '%c'\n", c);
    exit(1);
  }
}

void set_words(uint32_t *x, const char *hex_string, uint32_t count) {
  int index=0, length=0, value;

  for(index=0;index<count;index++)
    x[index]=0;
  while(hex_string[length]!=0)
    length++;
  for(index=0;index<length;index++) {
    value=nibble(hex_string[length-index-1]);
    x[index/8] += value<<index%8*4;
  }
}

int compare_words(uint32_t *x, uint32_t *y, uint32_t count) {
  int index;

  for(index=count-1;index>=0;index--) {
    if(x[index]!=y[index]) {
      if(x[index]>y[index])
        return 1;
      else
        return -1;
    }
  }
  return 0;
}

void swap_words(uint32_t *a, uint32_t *b, uint32_t count) {
  uint32_t temp;
  int      index;

  for(index=0;index<count;index++) {
    temp=a[index];
    a[index]=b[index];
    b[index]=temp;
  }
}

void random_words(uint32_t *x, uint32_t count) {
  int index;

  for(index=0;index<count;index++)
    x[index]=random_word();
}
