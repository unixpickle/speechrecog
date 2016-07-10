function [mat] = even_odd(size)
  mat = sparse(size, size);
  for i = 1:(size/2)
    mat(i, i*2-1) = 1;
    mat(i+size/2, i*2) = 1;
  end
end
